package target

import (
	"fmt"
	"github.com/chrismarget/cisco-l2t/attribute"
	"github.com/chrismarget/cisco-l2t/message"
	"github.com/chrismarget/cisco-l2t/rudp"
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
			ourIp, err := rudp.GetOutgoingIpForDestination(ip)
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
			ourIp, err := rudp.GetOutgoingIpForDestination(replyAddr)
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
// packets. It writes to a boolean channel whenever it's time to send a packet
// (true) or give up (false).
//
// With an initialRTTGuess of 100ms, a retryMultiplier of 2, and an end at
// 2500ms from start, the progression of writes to the channel would look like:
//  time    attempt channel
//  @t=0    0       true     (first packet is instant)
//  @t=100  1       true     (retransmit after 100ms)
//  @t=300  2       true     (retransmit after 200ms)
//  @t=700  3       true     (retransmit after 400ms)
//  @t=1500 4       true     (retransmit after 800ms)
//  @t=2500 -       -        (nothing happens at 'end')
//  @t=3100 -       false    (we're past 'end' and the retransmit timer expired)
func packetTimerFunc(doSend chan<- bool, end time.Time) {
	// initialize timers
	duration := initialRTTGuess

	// first packet should be sent immediately
	doSend <- true

	// loop until end time, progressively increasing the interval
	for time.Now().Before(end) {
		time.Sleep(duration)
		if time.Now().Before(end) {
			// there's still time left...
			doSend <- true
		} else {
			// timer expired while we were sleeping
			doSend <- false
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
	ourIp, err := rudp.GetOutgoingIpForDestination(destination.IP)
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

	out := rudp.SendThis{
		Payload:         payload,
		Destination:     destination,
		ExpectReplyFrom: nil,
		RttGuess:        initialRTTGuess,
	}

	in := rudp.Communicate(out, nil)
	if in.Err != nil {
		return testPacketResult{
			err: err,
		}
	}

	msg, err := message.UnmarshalMessage(in.ReplyData)
	if err != nil {
		return testPacketResult{
			err: err,
		}
	}

	err = msg.Validate()
	if err != nil {
		return testPacketResult{
			err: err,
		}
	}

	// Pull the name and platform strings from the message.
	var name string
	var platform string
	for t, a := range msg.Attributes() {
		if t == attribute.DevNameType {
			name = a.String()
		}
		if t == attribute.DevTypeType{
			platform = a.String()
		}
	}

	return testPacketResult{
		err:      in.Err,
		latency:  in.Rtt,
		IP:       in.ReplyFrom,
		platform: platform,
		name:     name,
	}

}

func initialLatency() []time.Duration {
	var l []time.Duration
	for len(l) < 5 {
		l = append(l, initialRTTGuess)
	}
	return l
}
