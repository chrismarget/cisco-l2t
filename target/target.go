package target

import (
	"bytes"
	"fmt"
	"github.com/chrismarget/cisco-l2t/attribute"
	"github.com/chrismarget/cisco-l2t/message"
	"log"
	"net"
	"strconv"
	"time"
)

const (
	udpPort           = 2228
	UdpProtocol       = "udp4"
	IPv4              = "ipv4"
	nilIP             = "<nil>"
	inBufferSize      = 65535
	initialRTTGuess   = 17 * time.Millisecond
	maxLatencySamples = 10
	maxRetries        = 10
	maxRTT            = 2500 * time.Millisecond
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
	useDial         bool
	latency         []time.Duration
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
func (o *defaultTarget) communicateViaConventionalSocket(b []byte) ([]byte, error) {
	var rtt time.Duration
	destination := &net.UDPAddr{
		IP:   o.theirIp[o.talkToThemIdx],
		Port: udpPort,
	}

	cxn, err := net.ListenUDP(UdpProtocol, &net.UDPAddr{IP: o.ourIp})
	if err != nil {
		return nil, err
	}
	defer cxn.Close()


	// todo: there's no retry, no timeout here yet
	buffIn := make([]byte, inBufferSize)

	received := 0
	retries := 0
	for received == 0 {
		if retries >= maxRetries {
			return nil, fmt.Errorf("Lost connection with switch %s after %d retries", destination.IP.String(), retries)
		}
		wait := o.estimateLatency()
		var respondent *net.UDPAddr

		// Send the packet
		n, err := cxn.WriteToUDP(b, destination)
		switch {
		case err != nil:
			return nil, err
		case n != len(b):
			return nil, fmt.Errorf("attemtped send of %d bytes, only managed %d", len(b), n)
		}

		// collect start time for later RTT calculation, set deadline
		start := time.Now()
		err = cxn.SetReadDeadline(start.Add(wait))
		if err != nil {
			return nil, err
		}

		// read until packet or deadline
		received, respondent, err = cxn.ReadFromUDP(buffIn)
		rtt = time.Since(start)
		// How can things go wrong here?
		switch {
		case err != nil:
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// Socket timeout; Double the timeout interval, stick it in the latency history.
				retries += 1
				o.updateLatency(2 * rtt)
				continue
			} else {
				// Mystery error
				return nil, fmt.Errorf("error waiting for reply via conventional socket - %s", err.Error())
			}
		case n == len(buffIn):
			// Unexpectedly large read
			return nil, fmt.Errorf("got full buffer: %d bytes", n)
		case !respondent.IP.Equal(o.theirIp[o.listenToThemIdx]):
			// Alien reply
			tIp := o.theirIp[o.talkToThemIdx].String()
			eIp := o.theirIp[o.listenToThemIdx].String()
			aIp := respondent.IP.String()
			return nil, fmt.Errorf("%s replied from unexpected address %s, rather than %s", tIp, aIp, eIp)
		}
	}
	o.updateLatency(rtt)
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

// estimateLatency tries to estimate the response time for this target
// using the contents of the objects latency slice.
func (o *defaultTarget) estimateLatency() time.Duration {
	if len(o.latency) == 0 {
		return initialRTTGuess
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

// updateLatency adds the passed time.Duration as the most recent
// latency sample, trims the latency slice to size.
func (o *defaultTarget) updateLatency(t time.Duration) {
	log.Println("samples: ",o.latency)
	log.Println("update:", t)
	o.latency = append(o.latency, t)
	log.Println("new samples: ",o.latency)
	// delete old elements of latency slice because we care more about
	// recent data (and certainly want to purge early bad assumptions)
	if len(o.latency) > maxLatencySamples {
		log.Println("trim")
		o.latency = o.latency[len(o.latency)-maxLatencySamples : len(o.latency)]
	}

}

type SendMessageConfig struct {
	M     message.Msg
	Inbox chan MessageResponse
}

type MessageResponse struct {
	Response message.Msg
	Err      error
}
