package target

import (
	"fmt"
	"net"
	"time"
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
	wait := initialRTTGuess
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
