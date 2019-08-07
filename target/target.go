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
	nilIP             = "<nil>"
	inBufferSize      = 65535
	initialRTTGuess   = 250 * time.Millisecond
	maxLatencySamples = 10
	maxRetries        = 10
	maxRTT            = 2500 * time.Millisecond
	retryMultiplier   = 2
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
	name            string
	platform        string
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

	out.WriteString("\nTarget Hostname:    ")
	switch o.name {
	case "":
		out.WriteString("<unknown>")
	default:
		out.WriteString(o.name)
	}

	out.WriteString("\nTarget Platform:    ")
	switch o.platform {
	case "":
		out.WriteString("<unknown>")
	default:
		out.WriteString(o.platform)
	}

	out.WriteString("\nKnown IP Addresses:")
	for _, ip := range o.theirIp {
		out.WriteString(" ")
		out.WriteString(ip.String())
	}

	out.WriteString("\nTarget address:     ")
	switch {
	case o.talkToThemIdx >= 0:
		out.WriteString(o.theirIp[o.talkToThemIdx].String())
	default:
		out.WriteString("none")
	}

	out.WriteString("\nListen address:     ")
	switch {
	case o.listenToThemIdx >= 0:
		out.WriteString(o.theirIp[o.listenToThemIdx].String())
	default:
		out.WriteString("none")
	}

	out.WriteString("\nLocal address:      ")
	switch o.ourIp.String() {
	case nilIP:
		out.WriteString("none")
	default:
		out.WriteString(o.ourIp.String())
	}

	out.WriteString("\nUse Dial:           ")
	out.WriteString(strconv.FormatBool(o.useDial))

	return out.String()
}

// communicateViaConventionalSocket sends a byte slice to the target using a UDP
// socket. It includes retry logic for handling packet loss. On success it
// returns a byte slice containing the reply payload.
//
// This function is useful for targets that reply from an address other
// than the one we talk to. In an ideal world only the communicateViaDialSocket
// method would be required, but some firewall/NAT situations may force the use
// of this method.
func (o *defaultTarget) communicateViaConventionalSocket(b []byte) ([]byte, error) {
	destination := &net.UDPAddr{
		IP:   o.theirIp[o.talkToThemIdx],
		Port: udpPort,
	}

	cxn, err := net.ListenUDP(UdpProtocol, &net.UDPAddr{IP: o.ourIp})
	if err != nil {
		return nil, err
	}
	defer cxn.Close()

	buffIn := make([]byte, inBufferSize)

	// Collect start time for later RTT calculation. Note that we're only
	// collecting the start time *once* even though we may send several
	// packets. In the case of no jitter/loss, the numbers will be correct.
	// In case we send the packet twice, which measurement is correct?
	// Elapsed time since packet zero or since packet one? We're
	// deliberately opting to accept the more pessimistic measurement.
	start := time.Now()
	var rtt time.Duration
	received := 0
	attempts := 0
	for received == 0 {
		if attempts >= maxRetries {
			return nil, fmt.Errorf("lost connection with switch %s after %d attempts", destination.IP.String(), attempts)
		}
		wait := o.estimateLatency()

		// Send the packet. Error handling happens after noting the start time.
		n, err := cxn.WriteToUDP(b, destination)

		switch {
		case err != nil:
			return nil, err
		case n != len(b):
			return nil, fmt.Errorf("attemtped send of %d bytes, only managed %d", len(b), n)
		}

		// set deadline based on start time
		err = cxn.SetReadDeadline(start.Add(wait))
		if err != nil {
			return nil, err
		}

		// read until packet or deadline
		var respondent *net.UDPAddr
		received, respondent, err = cxn.ReadFromUDP(buffIn)

		// Note the elapsed time
		rtt = time.Since(start)

		// How might things have gone wrong?
		switch {
		case err != nil:
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// Socket timeout; Double the timeout interval, stick it in the latency history.
				attempts += 1
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
			// Alien reply. Nudge the latency budget, try again.
			o.updateLatency(wait)
			continue
		}
	}
	o.updateLatency(rtt)
	return buffIn, nil
}

// communicateViaDialSocket sends a byte slice to the target using a UDP
// socket. It includes retry logic for handling packet loss. On success it
// returns a byte slice containing the reply payload.
//
// Because this function uses the net.DialUDP method it is only useful for
// targets that reply from the address to which we sent the query. That is the
// natural behavior of most UDP services, but not the Cisco L2T server. This is
// the preferred method because it's friendly to stateful middleboxes like NAT.
func (o *defaultTarget) communicateViaDialSocket(b []byte) ([]byte, error) {
	destination := &net.UDPAddr{
		IP:   o.theirIp[o.talkToThemIdx],
		Port: udpPort,
	}

	cxn, err := net.DialUDP(UdpProtocol, &net.UDPAddr{}, destination)
	if err != nil {
		return nil, err
	}
	// todo: this close causes ICMP unreachables. some sort of delay where the socket
	//  hangs around after we don't need it anymore would be good.
	defer cxn.Close()

	buffIn := make([]byte, inBufferSize)

	var rtt time.Duration
	received := 0
	attempts := 0
	for received == 0 {
		if attempts >= maxRetries {
			return nil, fmt.Errorf("lost connection with switch %s after %d attempts", destination.IP.String(), attempts)
		}
		wait := o.estimateLatency()
		n, err := cxn.Write(b)
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
		log.Println(wait)

		// read until packet or deadline
		received, err = cxn.Read(buffIn)
		rtt = time.Since(start)

		// How might things have gone wrong?
		switch {
		case err != nil:
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// Socket timeout; Double the timeout interval, stick it in the latency history.
				attempts += 1
				o.updateLatency(2 * rtt)
				continue
			} else {
				// Mystery error
				return nil, fmt.Errorf("error waiting for reply via conventional socket - %s", err.Error())
			}
		case n == len(buffIn):
			// Unexpectedly large read
			return nil, fmt.Errorf("got full buffer: %d bytes", n)
		}
	}

	o.updateLatency(rtt)
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
	o.latency = append(o.latency, t)
	// delete old elements of latency slice because we care more about
	// recent data (and certainly want to purge early bad assumptions)
	if len(o.latency) > maxLatencySamples {
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
