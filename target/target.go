package target

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/chrismarget/cisco-l2t/message"
)

const (
	udpPort      = 2228
	UdpProtocol  = "udp4"
	IPv4         = "ipv4"
	nilIP        = "<nil>"
	inBufferSize = 2048
	initialRTT   = 17 * time.Millisecond
	maxRTT       = 2500 * time.Millisecond
)

var (
	testMsg = []byte{
		1, 1, 0, 31, 4, 2, 8, 255, 255, 255, 255, 255, 255,
		1, 8, 255, 255, 255, 255, 255, 255, 3, 4, 0, 4, 14, 6,
	}
)

type Target interface {
	String() string
}

type defaultTarget struct {
	theirIp         []net.IP
	talkToThemIdx   int
	listenToThemIdx int
	cxn             *net.UDPConn
	ourIp           net.IP
	useDial         bool
	outgoing        chan SendMessageConfig
}

func (o *defaultTarget) Send(m SendMessageConfig) {
	o.outgoing <- m
}

func (o defaultTarget) String() string {
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
	if addressIsNew(&ip, o.addresses) {
		o.addresses = append(o.addresses, ip)
	}
	return o
}

func (o *defaultTargetBuilder) Build() (Target, error) {
	// TODO: plz initialize as close to 'return' as possible.
	result := defaultTarget{
		theirIp:         o.addresses,
		talkToThemIdx:   -1,
		listenToThemIdx: -1,
		outgoing:        make(chan SendMessageConfig),
	}

	// We keep track of every address the target replies from, ordered for
	// correlation with the slice of addresses we spoke to (o.addresses)
	var observedIps []net.IP

	// Loop until every element in result.theirIp
	// has a corresponding observedIps element
	for len(result.theirIp) > len(observedIps) {
		testIp := result.theirIp[len(observedIps)]
		respondingIp, err := checkTargetIp(&testIp)
		if err != nil {
			return nil, err
		}

		if respondingIp != nil {
			// We got a reply. Add the address to the list(s) as appropriate.
			observedIps = append(observedIps, *respondingIp)
			if addressIsNew(respondingIp, o.addresses) {
				o.addresses = append(o.addresses, *respondingIp)
			}
		} else {
			// No reply, dump a placeholder in the list.
			observedIps = append(observedIps, nil)
		}
	}

	// Check to see whether we had symmetric comms with any of
	// those addresses. If so, open a connection.
	for i, ip := range result.theirIp {
		//if ip.String() == observedIps[i].String() {
		if ip.Equal(observedIps[i]) {
			cxn, err := net.DialUDP(UdpProtocol, nil, &net.UDPAddr{IP: ip, Port: udpPort})
			if err != nil {
				return result, err
			}
			// Get the local system address that faces that target.
			ourIp, err := getSrcIp(&ip)

			result.ourIp = *ourIp
			result.talkToThemIdx = i
			result.listenToThemIdx = i
			result.useDial = true
			result.cxn = cxn
			spawnSenderRoutine(result.outgoing, cxn)
			return result, nil
		}
	}

	// If we got here, then no symmetric comms are possible.
	// Did we get ANY reply?
	// If so, open a listener for use with this target.

	// Loop over observed (reply source) address list. Ignore any that are <nil>
	for i, replyAddr := range observedIps {
		if replyAddr != nil {
			// We found one. The target replies from "replyAddr".
			// Get the local system address that faces that target.
			ourIp, err := getSrcIp(&replyAddr)
			result.ourIp = *ourIp

			// Record the index of the address we need to talk to.
			result.talkToThemIdx = i

			// Find the responding address in the target's address list.
			for k, v := range result.theirIp {
				if v.String() == replyAddr.String() {
					result.listenToThemIdx = k
				}
			}

			cxn, err := net.ListenUDP(UdpProtocol, &net.UDPAddr{IP: *ourIp})
			if err != nil {
				return result, err
			}

			result.cxn = cxn
		}
	}

	spawnSenderRoutine(result.outgoing, result.cxn)
	return result, nil
}

// TODO: Carefully consider where this code should go / how it's invoked.
//  What happens if a caller calls 'Build()' more than once?
//  What happens if Send() is called and this thread has not been spawned yet?
func spawnSenderRoutine(messagesToSend chan SendMessageConfig, c *net.UDPConn) {
	go func() {
		for sendConfig := range messagesToSend {
			// TODO: Need a timeout.
			_, err := c.Write(sendConfig.M.Marshal(nil))
			if err != nil {
				sendConfig.Inbox <- MessageResponse{
					Err: err,
				}
				continue
			}

			// TODO: Fixed buffer size == not good
			// TODO: Need a timeout.
			b := make([]byte, inBufferSize)
			_, err = c.Read(b)
			if err != nil {
				sendConfig.Inbox <- MessageResponse{
					Err: err,
				}
				continue
			}

			m, err := message.UnmarshalMessage(b)
			if err != nil {
				sendConfig.Inbox <- MessageResponse{
					Err: err,
				}
				continue
			}

			sendConfig.Inbox <- MessageResponse{
				Response: m,
			}
		}

		// TODO: If the channel is closed and the loop exits, should
		//  this routine close the socket or cleanup other stuff?
	}()
}

func NewTarget() Builder {
	return &defaultTargetBuilder{}
}

// getSrcIp returns a *net.IP representing the local interface
// that's best suited for talking to the passed target address
func getSrcIp(t *net.IP) (*net.IP, error) {
	c, err := net.Dial("udp4", t.String()+":1")
	if err != nil {
		return nil, err
	}
	defer c.Close()

	return &c.LocalAddr().(*net.UDPAddr).IP, nil
}

// testTarget sends a test L2T message to the specified IP address. It
// returns the address that replied to the message without evaluating
// the contents of the reply, <nil> if no reply.
func checkTargetIp(t *net.IP) (*net.IP, error) {
	ourIp, err := getSrcIp(t)
	if err != nil {
		return nil, err
	}
	payload := append(testMsg, *ourIp...)

	conn, err := net.ListenUDP(UdpProtocol, &net.UDPAddr{})
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	timedOut := false
	wait := initialRTT
	buffIn := make([]byte, inBufferSize)
	var respondent *net.UDPAddr
	for timedOut == false {
		log.Println("sending to", t)
		if wait > maxRTT {
			timedOut = true
			wait = maxRTT
		}

		n, err := conn.WriteToUDP(payload, &net.UDPAddr{IP: *t, Port: udpPort})
		if err != nil {
			return nil, err
		}
		if n != len(payload) {
			return nil, fmt.Errorf("attemtped send of %d bytes, only managed %d", len(payload), n)
		}

		err = conn.SetReadDeadline(time.Now().Add(wait))
		if err != nil {
			return nil, err
		}

		n, respondent, err = conn.ReadFromUDP(buffIn)
		if n == len(buffIn) {
			return nil, fmt.Errorf("got full buffer: %d bytes", n)
		}

		if n > 0 {
			break
		} else {
			wait = wait * 4
		}
	}

	if timedOut == true {
		return nil, nil
	}

	return &respondent.IP, nil
}

// addressIsNew returns a boolean indicating whether
// the net.IP is found in the []net.IP
func addressIsNew(a *net.IP, known []net.IP) bool {
	for _, k := range known {
		if a.String() == k.String() {
			return false
		}
	}
	return true
}
