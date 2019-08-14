package target

import (
	"fmt"
	"github.com/chrismarget/cisco-l2t/attribute"
	"github.com/chrismarget/cisco-l2t/message"
	"github.com/chrismarget/cisco-l2t/communicate"
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
			return nil, fmt.Errorf("error checking %s - %s", result.IP, result.err.Error())
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
			ourIp, err := communicate.GetOutgoingIpForDestination(ip)
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
			ourIp, err := communicate.GetOutgoingIpForDestination(replyAddr)
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
		Port: communicate.CiscoL2TPort,
	}

	// Build up the test message. Doing so requires that we know our IP address
	// which, on a multihomed system requires that we look up the route to the
	// target. So, we need to know about the target before we can form the
	// message.
	ourIp, err := communicate.GetOutgoingIpForDestination(destination.IP)
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

	// We're going to send the message via two different sockets: A "connected"
	// (dial) socket and a "non-connected" (listen) socket. The former can
	// telegraph ICMP unreachable (go away!) messages to us, while the latter
	// can detect 3rd party replies (necessary because of course the Cisco L2T
	// service generates replies from an alien (NAT unfriendly!) address.
	stopDialSocket := make(chan struct{}) // abort channel
	outViaDial := communicate.SendThis{ // Communicate() output structure
		Payload:         payload,
		Destination:     destination,
		ExpectReplyFrom: destination.IP,
		RttGuess:        communicate.InitialRTTGuess,
	}
	stopListenSocket := make(chan struct{}) // abort channel
	outViaListen := communicate.SendThis{ // Communicate() output structure
		Payload:         payload,
		Destination:     destination,
		ExpectReplyFrom: nil,
		RttGuess:        communicate.InitialRTTGuess,
	}

	dialResult := make(chan communicate.SendResult)
	go func() {
		dialResult <- communicate.Communicate(outViaDial, stopDialSocket)
	}()

	listenResult := make(chan communicate.SendResult)
	go func() {
		// This guy can't hear ICMP unreachables, so keep the noise down
		// by starting him a bit after the "dial" based listener.
		time.Sleep(communicate.InitialRTTGuess)
		listenResult <- communicate.Communicate(outViaListen, stopListenSocket)
	}()

	// grab a SendResult from either channel (socket)
	var in communicate.SendResult
	select {
	case in = <-dialResult:
	case in = <-listenResult:
	}
	close(stopDialSocket)
	close(stopListenSocket)

	// return an error (maybe)
	if in.Err != nil {
		if result, ok := in.Err.(net.Error); ok && result.Timeout() {
			// we timed out. Return an empty testPacketResult
			return testPacketResult{}
		} else {
			// some other type of error
			return testPacketResult{err: in.Err}
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
		if t == attribute.DevTypeType {
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
		l = append(l, communicate.InitialRTTGuess)
	}
	return l
}
