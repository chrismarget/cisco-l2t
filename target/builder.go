package target

import (
	"github.com/chrismarget/cisco-l2t/attribute"
	"github.com/chrismarget/cisco-l2t/communicate"
	"github.com/chrismarget/cisco-l2t/message"
	"net"
	"strings"
	"sync"
	"time"
)

type Builder interface {
	AddIp(net.IP) Builder
	Build() (Target, error)
}

func TargetBuilder() Builder {
	return &defaultTargetBuilder{}
}

type targetInfo struct {
	destination *net.UDPAddr
	theirSource net.IP
	localAddr   net.IP
	rtt         []time.Duration
	bestRtt     time.Duration
}

type defaultTargetBuilder struct {
	addresses []net.IP
}

func (o *defaultTargetBuilder) AddIp(ip net.IP) Builder {
	if addressIsNew(ip, o.addresses) {
		o.addresses = append(o.addresses, ip)
	}
	return o
}

func (o *defaultTargetBuilder) Build() (Target, error) {
	var name string
	var platform string
	var mgmtIp net.IP
	var info []targetInfo

	// Loop over o.addresses, noting that it may grow as the loop progresses
	var i int
	for i < len(o.addresses) {
		destination := &net.UDPAddr{
			IP:   o.addresses[len(info)],
			Port: communicate.CiscoL2TPort,
		}
		result := checkTarget(destination)

		// Save "name" and "result" so they're not
		// overwritten by a future failed query.
		if result.name != "" {
			name = result.name
		}
		if result.platform != "" {
			platform = result.platform
		}
		if result.mgmtIp.String() != "<nil>" {
			mgmtIp = result.mgmtIp
		}

		var rttSamples []time.Duration
		if result.sourceIp != nil {
			rttSamples = append(rttSamples, result.latency)
		}
		// Add a targetInfo structure to the slice for every address we probe.
		info = append(info, targetInfo{
			localAddr:   result.localIp,
			destination: destination,
			theirSource: result.sourceIp,
			rtt:         rttSamples,
			bestRtt:     result.latency, // only sample is best sample
		})

		// Reply came from an unknown source address? Add it to the list.
		if result.sourceIp != nil && addressIsNew(result.sourceIp, o.addresses) {
			o.addresses = append(o.addresses, result.sourceIp)
		}

		// Reply cited an unknown management address? Add it to the list.
		if result.mgmtIp != nil && addressIsNew(result.mgmtIp, o.addresses) {
			o.addresses = append(o.addresses, result.mgmtIp)
		}

		// Last iteration of this loop is the one where 'i' has grown to match
		// the length of the address list.
		i++
	}

	// look through the targetInfo structures we've collected
	var fastestTarget int
	var reachable bool
	for i, ti := range info {
		// ignore targetInfo if the target didn't talk to us
		if ti.theirSource == nil {
			continue
		} else {
			reachable = true
		}

		if ti.bestRtt < info[fastestTarget].bestRtt {
			fastestTarget = i
		}
	}

	return &defaultTarget{
		reachable: reachable,
		info:      info,
		best:      fastestTarget,
		name:      name,
		platform:  platform,
		mgmtIp:    mgmtIp,
	}, nil
}

func TestTargetBuilder() Builder {
	return &testTargetBuilder{}
}

type testTargetBuilder struct {
	addresses []net.IP
}

func (o *testTargetBuilder) AddIp(ip net.IP) Builder {
	if addressIsNew(ip, o.addresses) {
		o.addresses = append(o.addresses, ip)
	}
	return o

}
func (o *testTargetBuilder) Build() (Target, error) {
	name := "TestTarget"
	platform := "TestPlatform"
	mgmtIp := net.ParseIP("192.168.255.1")
	var ti []targetInfo
	if len(o.addresses) == 0 {
		o.addresses = append(o.addresses, net.ParseIP("127.0.0.1"))
	}

	for i, a := range o.addresses {
		outIp, _ := communicate.GetOutgoingIpForDestination(a)
		rtt := []time.Duration{(time.Duration(i) + 1) * time.Millisecond}
		ti = append(ti, targetInfo{
			destination: &net.UDPAddr{
				IP:   a,
				Port: communicate.CiscoL2TPort,
			},
			theirSource: a,
			localAddr:   outIp,
			rtt:         rtt,
			bestRtt:     0,
		})
	}

	return &defaultTarget{
		reachable: true,
		info:      ti,
		best:      0,
		name:      name,
		platform:  platform,
		mgmtIp:    mgmtIp,
		rttLock:   sync.Mutex{},
	}, nil

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
	destination *net.UDPAddr
	err         error
	latency     time.Duration
	sourceIp    net.IP
	platform    string
	name        string
	mgmtIp      net.IP
	localIp     net.IP
}

func (r *testPacketResult) String() string {
	var s strings.Builder
	s.WriteString("result:\n  ")
	s.WriteString(r.destination.String())
	s.WriteString("\n  ")
	s.WriteString(r.sourceIp.String())
	s.WriteString("\n")
	return s.String()
}

// checkTarget sends test L2T messages to the specified IP address. It
// returns a testPacketResult that represents the result of the check.
func checkTarget(destination *net.UDPAddr) testPacketResult {
	// Build up the test message. Doing so requires that we know our IP address
	// which, on a multihomed system requires that we look up the route to the
	// target. So, we need to know about the target before we can form the
	// message.
	ourIp, err := communicate.GetOutgoingIpForDestination(destination.IP)
	if err != nil {
		return testPacketResult{
			destination: destination,
			err:         err,
		}
	}
	ourIpAttr, err := attribute.NewAttrBuilder().
		SetType(attribute.SrcIPv4Type).
		SetString(ourIp.String()).
		Build()
	if err != nil {
		return testPacketResult{
			err: err,
		}
	}

	testMsg, err := message.TestMsg()
	if err != nil {
		return testPacketResult{err: err}
	}

	err = testMsg.Validate()
	if err != nil {
		return testPacketResult{err: err}
	}

	payload := testMsg.Marshal([]attribute.Attribute{ourIpAttr})

	// We're going to send the message via two different sockets: A "connected"
	// (dial) socket and a "non-connected" (listen) socket. The former can
	// telegraph ICMP unreachable (go away!) messages to us, while the latter
	// can detect 3rd party replies (necessary because of course the Cisco L2T
	// service generates replies from an alien (NAT unfriendly!) address.
	stopDialSocket := make(chan struct{}) // abort channel
	outViaDial := communicate.SendThis{   // Communicate() output structure
		Payload:         payload,
		Destination:     destination,
		ExpectReplyFrom: destination.IP,
		RttGuess:        communicate.InitialRTTGuess * 2,
	}
	stopListenSocket := make(chan struct{}) // abort channel
	outViaListen := communicate.SendThis{   // Communicate() output structure
		Payload:         payload,
		Destination:     destination,
		ExpectReplyFrom: nil,
		RttGuess:        communicate.InitialRTTGuess * 2,
	}

	dialResult := make(chan communicate.SendResult)
	go func() {
		// This guy can't hear 3rd party (alien) replies. Start him first
		// because he's not deaf to (noisy) ICMP unreachables.
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

	replyMsg, err := message.UnmarshalMessage(in.ReplyData)
	if err != nil {
		return testPacketResult{err: err}
	}

	err = replyMsg.Validate()
	if err != nil {
		return testPacketResult{err: err}
	}

	// Pull the name, platform, and IP address from the message attributes.
	var name string
	var platform string
	var mgmtIp net.IP
	for t, a := range replyMsg.Attributes() {
		if t == attribute.DevNameType {
			name = a.String()
		}
		if t == attribute.DevTypeType {
			platform = a.String()
		}
		if t == attribute.DevIPv4Type {
			mgmtIp = net.ParseIP(a.String())
		}
	}

	return testPacketResult{
		localIp:  ourIp,
		err:      in.Err,
		latency:  in.Rtt,
		sourceIp: in.ReplyFrom,
		platform: platform,
		name:     name,
		mgmtIp:   mgmtIp,
	}
}

func initialLatency() []time.Duration {
	var l []time.Duration
	for len(l) < 5 {
		l = append(l, communicate.InitialRTTGuess)
	}
	return l
}
