package target

import (
	"fmt"
	"github.com/chrismarget/cisco-l2t/attribute"
	"github.com/chrismarget/cisco-l2t/message"
	"net"
	"time"
)

const (
	timesUp = -1
)

// todo: consider using message.TestMsg instead?
var (
	testMsg = []byte{
		1, 1, 0, 31, 4, 2, 8, 255, 255, 255, 255, 255, 255,
		1, 8, 255, 255, 255, 255, 255, 255, 3, 4, 0, 1, 14, 6,
	}
)

type Builder interface {
	AddIp(net.IP) Builder
	Build() (Target, error)
}

func TargetBuilder() Builder {
	return &defaultTargetBuilder{}
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
	var result testPacketResult
	var name string
	var platform string

	// Loop until every element in o.addresses has a corresponding
	// observedIps element (these will be <nil> if no reply)
	for len(o.addresses) > len(observedIps) {
		testIp := o.addresses[len(observedIps)]
		result = checkTargetIp(testIp)
		if result.err != nil {
			return nil, fmt.Errorf("no response from address %s - %s", result.IP, result.err.Error())
		}

		// save "name" and "result" so they're not crushed by a future failed query
		if result.name != "" {
			name = result.name
		}
		if result.platform != "" {
			platform = result.platform
		}

		// add the result (maybe <nil>) to the list of observed addresses
		observedIps = append(observedIps, result.IP)

		// Did we hear back from the target?
		if result.IP != nil {
			latency = append(latency, result.latency)
			// if the observed address is previously unseen, add it to o.addresses
			if addressIsNew(result.IP, o.addresses) {
				o.addresses = append(o.addresses, result.IP)
			}
		}
	}

	// Now that we've probed every address, check to see whether we had
	// symmetric comms with any of those addresses. These will be index
	// locations where o.addresses and observedIps have the same value.
	// todo: we could be inspecting RTT latency here to make an even better choice
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
				latency:         initialLatency(),
				name:            name,
				platform:        platform,
			}, nil
		}
	}

	// If we got here, then no symmetric comms are possible.
	// Did we get ANY reply?

	// Loop over observed (reply source) address list. Ignore any that are <nil>
	// todo: we could be inspecting RTT latency here to make an even better choice
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
				useDial:         false,
				latency:         initialLatency(),
				name:            name,
				platform:        platform,
			}, nil
		}
	}

	return nil, &UnreachableTargetError{
		AddressesTried: o.addresses,
	}
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

// packetTimerFunc is used to coordinate (re)transmission of unreliable UDP
// packets. It writes a progression of integers represening attempt numbers
// to the counter channel. It sends timesUp (-1) when the timer expires. sfox
// probably hates that the retry attempt value is overloaded in this way :)
//
// With an initialRTTGuess of 100ms and a retryMultiplier of 2, the
// progression of writes to the channel would look like:
//  @t=0    0 (first packet is instant)
//  @t=100  1 (retransmit after 100ms)
//  @t=300  2 (retransmit after 200ms)
//  @t=700  3 (retransmit after 400ms)
//  @t=1500 4 (retransmit after 800ms)
func packetTimerFunc(counter chan<- int, end time.Time) {
	// initialize timers
	duration := initialRTTGuess

	// first iteration is written to the channel immediately
	iterations := 0
	counter <- iterations

	// loop until end time, progressively increasing the interval
	for time.Now().Before(end) {
		time.Sleep(duration)
		iterations++
		if time.Now().Before(end) {
			// there's still time left...
			counter <- iterations
		} else {
			// timer expired while we were sleeping
			counter <- timesUp
		}
		duration = duration * retryMultiplier
	}
}

type testPacketResult struct {
	err      error
	latency  time.Duration
	IP       net.IP
	platform string
	name     string
}

func sendFromNewSocket(result chan<- testPacketResult, payload []byte, destination *net.UDPAddr, end time.Time) {
	// create the socket
	conn, err := net.ListenUDP(UdpProtocol, &net.UDPAddr{})
	if err != nil {
		result <- testPacketResult{err: err}
		return
	}
	defer conn.Close()

	// send on the socket
	n, err := conn.WriteToUDP(payload, destination)
	start := time.Now()
	switch {
	case err != nil:
		result <- testPacketResult{err: err}
		return
	case n != len(payload):
		result <- testPacketResult{
			err: fmt.Errorf("attemtped send of %d bytes, only managed %d", len(payload), n),
		}
		return
	}

	// set the read deadline
	err = conn.SetReadDeadline(end)
	if err != nil {
		result <- testPacketResult{err: err}
		return
	}

	// read from the socket
	buffIn := make([]byte, inBufferSize)
	var respondent *net.UDPAddr
	n, respondent, err = conn.ReadFromUDP(buffIn)

	// Note the elapsed time
	rtt := time.Since(start)

	// How might things have gone wrong?
	switch {
	case err != nil:
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			// Socket timeout
			result <- testPacketResult{latency: 0}
		} else {
			// Mystery error
			result <- testPacketResult{err: err}
		}
	case n == len(buffIn):
		// Unexpectedly large read
		result <- testPacketResult{
			err: fmt.Errorf("got full buffer: %d bytes", n),
		}
	}

	// Unpack and and validate the message
	msg, err := message.UnmarshalMessage(buffIn)
	if err != nil {
		result <- testPacketResult{err: err}
		return
	}
	err = msg.Validate()
	if err != nil {
		result <- testPacketResult{err: err}
		return
	}

	var name string
	var platform string
	for _, att := range msg.Attributes() {
		if att.Type() == attribute.DevNameType {
			name = att.String()
		}
		if att.Type() == attribute.DevTypeType {
			platform = att.String()
		}
	}

	// If we got all the way here, it looks like we've got a legit reply.
	result <- testPacketResult{
		latency:  rtt,
		IP:       respondent.IP,
		name:     name,
		platform: platform,
	}

}

// checkTargetIp sends test L2T messages to the specified IP address. It
// returns a testPacketResult that represents the result of the check.
func checkTargetIp(target net.IP) testPacketResult {
	destination := &net.UDPAddr{
		IP:   target,
		Port: udpPort,
	}

	// Build up the test message. Doing so requires that we know our IP address
	// which, on a multihomed system requires that we look up the route to the
	// target. So, we need to know about the target before we can form the
	// message.
	ourIp, err := getOurIpForTarget(destination.IP)
	if err != nil {
		return testPacketResult{err: err}
	}
	ourIpAttr, err := attribute.NewAttrBuilder().
		SetType(attribute.SrcIPv4Type).
		SetString(ourIp.String()).
		Build()
	if err != nil {
		return testPacketResult{err: err}
	}

	payload := message.TestMsg().Marshal([]attribute.Attribute{ourIpAttr})

	// packetTimerFunc tells us when to send a packet (and the attempt number) here.
	sendNowChan := make(chan int)

	// sendFromNewSocket tells us what happened here.
	testResultChan := make(chan testPacketResult)

	// Start the timer that will tell us when to send packets
	end := time.Now().Add(maxRTT)
	go packetTimerFunc(sendNowChan, end)

	for {
		select {
		case attempt := <-sendNowChan:
			if attempt >= 0 {
				go sendFromNewSocket(testResultChan, payload, destination, end)
			} else {
				// timer expired. We never got a reply.
				return testPacketResult{latency: 0}
			}
		case testResult := <-testResultChan:
			return testResult
		}
	}
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

func initialLatency() []time.Duration {
	var l []time.Duration
	for len(l) < 5 {
		l = append(l, initialRTTGuess)
	}
	return l
}
