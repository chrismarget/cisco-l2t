package target

import (
	"bytes"
	"fmt"
	"github.com/chrismarget/cisco-l2t/attribute"
	"github.com/chrismarget/cisco-l2t/message"
	"net"
	"strconv"
	"time"
)

const (
	udpPort      = 2228
	UdpProtocol  = "udp4"
	IPv4         = "ipv4"
	nilIP        = "<nil>"
	inBufferSize = 65535
	initialRTT   = 17 * time.Millisecond
	maxRTT       = 2500 * time.Millisecond
	maxRetries   = 10
)

var (
	testMsg = []byte{
		1, 1, 0, 31, 4, 2, 8, 255, 255, 255, 255, 255, 255,
		1, 8, 255, 255, 255, 255, 255, 255, 3, 4, 0, 1, 14, 6,
	}
)

type Target interface {
	Send(message.Msg) (message.Msg, error)
	String() string
}

type defaultTarget struct {
	theirIp         []net.IP
	talkToThemIdx   int
	listenToThemIdx int
	ourIp           net.IP
	// todo: Why have I used two booleans here? Especially when this
	//  info can be divined from the talk/listenToThemIdx values
	useDial   bool
	useListen bool
	latency   []time.Duration
}

func (o *defaultTarget) Send(msg message.Msg) (message.Msg, error) {
	var payload []byte
	switch msg.NeedsSrcIp() {
	case true:
		srcIpAttr, err := attribute.NewAttrBuilder().
			SetType(attribute.SrcIPv4Type).
			SetString(o.ourIp.String()).
			Build()
		if err != nil {
			return nil, err
		}
		payload = msg.Marshal([]attribute.Attribute{srcIpAttr})
	case false:
		payload = msg.Marshal([]attribute.Attribute{})
	}

	switch o.useDial {
	case true:
		reply, err := o.communicateViaDialSocket(payload)
		if err != nil {
			return nil, err
		}
		return message.UnmarshalMessage(reply)
	case false:
		reply, err := o.communicateViaConventionalSocket(payload)
		if err != nil {
			return nil, err
		}
		return message.UnmarshalMessage(reply)
	}
	return nil, nil
}

func (o *defaultTarget) String() string {
	var out bytes.Buffer
	out.WriteString("Known IP Addresses:")
	for _, ip := range o.theirIp {
		out.WriteString(" ")
		out.WriteString(ip.String())
	}

	out.WriteString("\nTarget address: ")
	switch {
	case o.talkToThemIdx >= 0:
		out.WriteString(o.theirIp[o.talkToThemIdx].String())
	default:
		out.WriteString("none")
	}

	out.WriteString("\nListen address: ")
	switch {
	case o.listenToThemIdx >= 0:
		out.WriteString(o.theirIp[o.listenToThemIdx].String())
	default:
		out.WriteString("none")
	}

	out.WriteString("\nLocal address:  ")
	switch o.ourIp.String() {
	case nilIP:
		out.WriteString("none")
	default:
		out.WriteString(o.ourIp.String())
	}

	out.WriteString("\nUse Dial:       ")
	out.WriteString(strconv.FormatBool(o.useDial))

	return out.String()
}

//TODO timeout, retry, etc
func (o defaultTarget) communicateViaConventionalSocket(b []byte) ([]byte, error) {
	destination := &net.UDPAddr{
		IP:   o.theirIp[o.talkToThemIdx],
		Port: udpPort,
	}

	cxn, err := net.ListenUDP(UdpProtocol, &net.UDPAddr{IP: o.ourIp})
	if err != nil {
		return nil, err
	}
	defer cxn.Close()

	n, err := cxn.WriteToUDP(b, destination)
	switch {
	case err != nil:
		return nil, err
	case n != len(b):
		return nil, fmt.Errorf("attemtped send of %d bytes, only managed %d", len(b), n)
	}

	// todo: there's no retry, no timeout here yet
	buffIn := make([]byte, inBufferSize)

	received := 0
	wait := o.estimateLatency()
	start := time.Now()
	for received == 0 {
		err = cxn.SetReadDeadline(start.Add(wait))
		if err != nil {
			return nil, err
		}

		var respondent *net.UDPAddr
		received, respondent, err = cxn.ReadFromUDP(buffIn)
		rtt := time.Since(start)
		o.latency = append(o.latency, rtt)
		switch {
		case err != nil:
			return nil, err
		case n == len(buffIn):
			return nil, fmt.Errorf("got full buffer: %d bytes", n)
		case !respondent.IP.Equal(o.theirIp[o.listenToThemIdx]):
			tIp := o.theirIp[o.talkToThemIdx].String()
			eIp := o.theirIp[o.listenToThemIdx].String()
			aIp := respondent.IP.String()
			return nil, fmt.Errorf("%s replied from unexpected address %s, rather than %s", tIp, aIp, eIp)
		}
	}


	return buffIn, nil
}

// todo timeout, etc..
func (o defaultTarget) communicateViaDialSocket(b []byte) ([]byte, error) {
	destination := &net.UDPAddr{
		IP:   o.theirIp[o.talkToThemIdx],
		Port: udpPort,
	}

	cxn, err := net.DialUDP(UdpProtocol, &net.UDPAddr{}, destination)
	if err != nil {
		return nil, err
	}
	defer cxn.Close()

	buffIn := make([]byte, inBufferSize)
	//	start := time.Now()
	n, err := cxn.Write(b)
	switch {
	case err != nil:
		return nil, err
	case n != len(b):
		return nil, fmt.Errorf("attemtped send of %d bytes, only managed %d", len(b), n)
	}

	cxn.Read(buffIn)
	return buffIn, nil
}

// estimateLatency tries to estimate the response time for this target.
func (o *defaultTarget) estimateLatency() time.Duration {
	// delete old elements of latency slice because we care more about
	// recent data (and certainly want to purge early bad assumptions)
	if len(o.latency) > 10 {
		o.latency = o.latency[len(o.latency)-10 : len(o.latency)]
	}

	// short on samples? add some assumptions to the data
	for len(o.latency) <= 5 {
		o.latency = append(o.latency, 100*time.Millisecond)
	}

	var result int64
	for i, l := range o.latency {
		switch i {
		case 0:
			result = int64(l)
		default:
			result = (result + int64(l)) / 2
		}
	}
	return time.Duration(float32(result) * float32(1.15))
}

type SendMessageConfig struct {
	M     message.Msg
	Inbox chan MessageResponse
}

type MessageResponse struct {
	Response message.Msg
	Err      error
}



