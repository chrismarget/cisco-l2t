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
	}

	received, respondent, err := cxn.ReadFromUDP(buffIn)
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

type Builder interface {
	AddIp(net.IP) Builder
	Build() (Target, error)
}

type defaultTargetBuilder struct {
	addresses           []net.IP
	preferredAddressIdx int
}

func (o *defaultTargetBuilder) AddIp(ip net.IP) Builder {
	if addressIsNew(ip, o.addresses) {
		o.addresses = append(o.addresses, ip)
	}
	return o
}

func (o defaultTargetBuilder) Build() (Target, error) {
	// We keep track of every address the target replies from, ordered for
	// correlation with the slice of addresses we spoke to (o.addresses)
	var observedIps []net.IP
	var latency []time.Duration

	// Loop until every element in o.addresses has a corresponding
	// observedIps element (these will be <nil> if no reply)
	for len(o.addresses) > len(observedIps) {
		testIp := o.addresses[len(observedIps)]
		respondingIp, responseTime, err := checkTargetIp(testIp)
		if err != nil {
			return nil, fmt.Errorf("no response from address %s - %s", respondingIp, err.Error())
		}

		// add the result (maybe <nil>) to the list of observed addresses
		observedIps = append(observedIps, respondingIp)

		// Did we hear back from the target?
		if respondingIp != nil {
			latency = append(latency, responseTime)
			// if the observed address is previously unseen, add it to o.addresses
			if addressIsNew(respondingIp, o.addresses) {
				o.addresses = append(o.addresses, respondingIp)
			}
		}
	}

	// Now that we've probed every address, check to see whether we had
	// symmetric comms with any of those addresses. These will be index
	// locations where o.addresses and observedIps have the same value.
	for i, ip := range o.addresses {
		if ip.Equal(observedIps[i]) {
			// Get the local system address that faces that target.
			ourIp, err := getOurIpForTarget(ip)
			if err != nil {
				return nil, err
			}

			return &defaultTarget{
				theirIp:         o.addresses,
				talkToThemIdx:   i,
				listenToThemIdx: i,
				ourIp:           ourIp,
				useDial:         true,
			}, nil
		}
	}

	// If we got here, then no symmetric comms are possible.
	// Did we get ANY reply?

	// Loop over observed (reply source) address list. Ignore any that are <nil>
	for i, replyAddr := range observedIps {
		if replyAddr != nil {
			// We found one. The target replies from "replyAddr".
			// Get the local system address that faces that target.
			ourIp, err := getOurIpForTarget(replyAddr)
			if err != nil {
				return nil, err
			}

			// Find the responding address in the target's address list.
			var listenIdx int
			for k, v := range o.addresses {
				if v.String() == replyAddr.String() {
					listenIdx = k
				}
			}

			return &defaultTarget{
				theirIp:         o.addresses,
				talkToThemIdx:   i,
				listenToThemIdx: listenIdx,
				ourIp:           ourIp,
				useListen:       true,
			}, nil
		}
	}

	//return &defaultTarget{
	//	theirIp:         o.addresses,
	//},

	return nil, &UnreachableTargetError{
		AddressesTried: o.addresses,
	}
}

func NewTarget() Builder {
	return &defaultTargetBuilder{}
}

// getOurIpForTarget returns a *net.IP representing the local interface
// that's best suited for talking to the passed target address
func getOurIpForTarget(t net.IP) (net.IP, error) {
	c, err := net.Dial("udp4", t.String()+":1")
	if err != nil {
		return nil, err
	}
	defer c.Close()

	return c.LocalAddr().(*net.UDPAddr).IP, nil
}

// testTarget sends a test L2T message to the specified IP address. It
// returns the address that replied to the message without evaluating
// the contents of the reply, <nil> if no reply.
func checkTargetIp(t net.IP) (net.IP, time.Duration, error) {
	ourIp, err := getOurIpForTarget(t)
	if err != nil {
		return nil, 0, err
	}
	payload := append(testMsg, ourIp...)

	conn, err := net.ListenUDP(UdpProtocol, &net.UDPAddr{})
	if err != nil {
		return nil, 0, err
	}
	defer conn.Close()

	timedOut := false
	wait := initialRTT
	buffIn := make([]byte, inBufferSize)
	var latency time.Duration
	var respondent *net.UDPAddr
	for timedOut == false {
		if wait > maxRTT {
			timedOut = true
			wait = maxRTT
		}

		start := time.Now()
		n, err := conn.WriteToUDP(payload, &net.UDPAddr{IP: t, Port: udpPort})
		if err != nil {
			return nil, 0, err
		}
		if n != len(payload) {
			return nil, 0, fmt.Errorf("attemtped send of %d bytes, only managed %d", len(payload), n)
		}

		err = conn.SetReadDeadline(time.Now().Add(wait))
		if err != nil {
			return nil, 0, err
		}

		n, respondent, err = conn.ReadFromUDP(buffIn)
		stop := time.Now()
		latency = stop.Sub(start)
		if n == len(buffIn) {
			return nil, 0, fmt.Errorf("got full buffer: %d bytes", n)
		}

		if n > 0 {
			break
		} else {
			wait = wait * 4
		}
	}

	if timedOut == true {
		return nil, -1, nil
	}

	return respondent.IP, latency, nil
}

// addressIsNew returns a boolean indicating whether
// the net.IP is found in the []net.IP
func addressIsNew(a net.IP, known []net.IP) bool {
	for _, k := range known {
		if a.String() == k.String() {
			return false
		}
	}
	return true
}
