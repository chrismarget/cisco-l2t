package target

import (
	"fmt"
	"log"
	"net"
	"time"
)

const (
	timesUp = -1
)

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
				latency:         initialLatency(),
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
				useDial:         false,
				latency:         initialLatency(),
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
	log.Println(time.Now(), "-", end)

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
	err     error
	latency time.Duration
}

func getLatency(resultChan chan<- testPacketResult, target net.IP) {

	// build up the message we'll be using to test
	ourIp, err := getOurIpForTarget(target)
	if err != nil {
		resultChan <- testPacketResult{err: err}
	}
	payload := append(testMsg, ourIp...)

	// packetTimerFunc tells us when to send a packet (and the attempt number) here.
	sendNowChan := make(chan int)

	// sendFromNewSocket tells us what happened here.
	testResultChan := make(chan testPacketResult)
	_ = testResultChan

	// start the timer that will tell us when to send packets
	end := time.Now().Add(maxRTT)
	go packetTimerFunc(sendNowChan, end)

	for {
		select {
		case attempt := <-sendNowChan:
			if attempt >= 0 {
				go sendFromNewSocket(testResultChan, end, payload)
			} else {
				// timer expired. We never got a reply.
				break
			}
			//case testResult := <- testResultChan:

		}
	}

	//// Loop until expiration
	//for time.Now().Before(end) {
	//	attempts += 1
	//	wait := initialRTTGuess * time.Duration(attempts)
	//	go sendFromNewSocket(rtt, err)
	//
	//	attempts += 1
	//}

}

func sendFromNewSocket(result chan<- testPacketResult, end time.Time, payload []byte) {
	// create the socket
	conn, err := net.ListenUDP(UdpProtocol, &net.UDPAddr{})
	if err != nil {
		result <- testPacketResult{ err: err }
	}
	defer conn.Close()

	// send on the socket
	n, err := conn.WriteToUDP(payload, &net.UDPAddr{IP: t, Port: udpPort})
	// todo error handling
	switch {
	case err != nil:
		result <- testPacketResult{ err: err }
	case n != len(payload):
		result <- testPacketResult{
			err: fmt.Errorf("attemtped send of %d bytes, only managed %d", len(b), n),
		}
	}

}

// checkTargetIp sends test L2T messages to the specified IP address. It
// returns the address that replied to the message without evaluating
// the contents of the reply (<nil> if no reply) and the observed latency.
func checkTargetIp(t net.IP) (net.IP, time.Duration, error) {
	ourIp, err := getOurIpForTarget(t)
	if err != nil {
		return nil, 0, err
	}
	payload := append(testMsg, ourIp...)
	_ = payload

	resultChan := make(chan testPacketResult)

	go getLatency(resultChan, t)

	select {
	case <-resultChan:
	}

	//timedOut := false
	//wait := initialRTTGuess
	//buffIn := make([]byte, inBufferSize)
	//var xlatency time.Duration
	//var respondent *net.UDPAddr
	//for timedOut == false {
	//
	//	conn, err := net.ListenUDP(UdpProtocol, &net.UDPAddr{})
	//	if err != nil {
	//		return nil, 0, err
	//	}
	//	defer conn.Close()
	//
	//	if wait > maxRTT {
	//		timedOut = true
	//		wait = maxRTT
	//	}
	//
	//	// Send the packet. Error handling happens after noting the start time.
	//	n, err := conn.WriteToUDP(payload, &net.UDPAddr{IP: t, Port: udpPort})
	//
	//	// collect start time for later RTT calculation
	//	start := time.Now()
	//
	//	switch {
	//	case err != nil:
	//		return nil, err
	//	case n != len(payload):
	//		return nil, fmt.Errorf("attemtped send of %d bytes, only managed %d", len(b), n)
	//	}
	//
	//
	//	if err != nil {
	//		return nil, 0, err
	//	}
	//	if n != len(payload) {
	//		return nil, 0, fmt.Errorf("attemtped send of %d bytes, only managed %d", len(payload), n)
	//	}
	//
	//	err = conn.SetReadDeadline(time.Now().Add(wait))
	//	if err != nil {
	//		return nil, 0, err
	//	}
	//
	//	n, respondent, err = conn.ReadFromUDP(buffIn)
	//	stop := time.Now()
	//	xlatency = stop.Sub(start)
	//	if n == len(buffIn) {
	//		return nil, 0, fmt.Errorf("got full buffer: %d bytes", n)
	//	}
	//
	//	if n > 0 {
	//		break
	//	} else {
	//		wait = wait * 4
	//	}
	//}
	//
	//if timedOut == true {
	//	return nil, -1, nil
	//}
	//
	//return respondent.IP, xlatency, nil
	return nil, 0, nil
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
