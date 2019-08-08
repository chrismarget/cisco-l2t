package target

import (
	"fmt"
	"github.com/chrismarget/cisco-l2t/attribute"
	"github.com/chrismarget/cisco-l2t/message"
	"net"
	"time"
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
	if len(o.addresses) == 0 {
		return nil, fmt.Errorf("no target addresses were provided")
	}

	// We keep track of every address the target replies from, ordered for
	// correlation with the slice of addresses we spoke to.
	var prefIndex int
	var latency []time.Duration
	var name string
	var platform string
	var pref string
	var last string
	var f []failure
	lowestLatency := 1 * time.Minute
	observedIps := o.addresses
	// TODO: Use the map instead of a slice.
	good := make(map[string]net.IP)
	var goods []net.IP

	for i := 0; i < len(observedIps); i++ {
		testIp := observedIps[i]
		if _, ok := good[testIp.String()]; ok {
			continue
		}

		result := checkTargetIp(testIp)
		if result.err != nil {
			f = append(f, failure{
				ip:  testIp,
				err: result.err,
			})
			if len(f) == len(observedIps) {
				return nil, fmt.Errorf("failed to communicate with all provided addresses")
			}
			continue
		}

		if _, ok := good[result.IP.String()]; !ok {
			observedIps = append(observedIps, result.IP)
		}

		good[testIp.String()] = testIp
		goods = append(goods, testIp)

		// Check if we should prefer this address.
		if result.latency < lowestLatency {
			pref = testIp.String()
			lowestLatency = result.latency
			prefIndex = len(goods) - 1
		} else if testIp.Equal(result.IP) {
			pref = testIp.String()
			lowestLatency = result.latency
			prefIndex = len(goods) - 1
		}

		latency = append(latency, result.latency)

		if len(name) == 0 {
			name = result.name
		}

		if len(platform) == 0 {
			platform = result.platform
		}

		last = testIp.String()
	}

	if len(pref) == 0 {
		pref = last
		lowestLatency = initialRTTGuess
	}

	ourIp, err := getOurIpForTarget(good[pref])
	if err != nil {
		return nil, fmt.Errorf("failed to get local IP address for communicating with preffered address '%s' - %s",
			pref, err.Error())
	}

	return &defaultTarget{
		theirIp:         goods,
		talkToThemIdx:   prefIndex,
		listenToThemIdx: prefIndex,
		ourIp:           ourIp,
		latency:         latency,
		name:            name,
		platform:        platform,
	}, nil
}

type failure struct {
	ip  net.IP
	err error
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

func sendFromNewSocket(payload []byte, destination *net.UDPAddr, end time.Time) testPacketResult {
	// create the socket
	conn, err := net.ListenUDP(UdpProtocol, &net.UDPAddr{})
	if err != nil {
		return testPacketResult{err:err}
	}
	defer conn.Close()

	// send on the socket
	n, err := conn.WriteToUDP(payload, destination)
	start := time.Now()
	switch {
	case err != nil:
		return testPacketResult{err:err}
	case n != len(payload):
		return testPacketResult{
			err: fmt.Errorf("attemtped send of %d bytes, only managed %d", len(payload), n),
		}
	}

	// set the read deadline
	err = conn.SetReadDeadline(end)
	if err != nil {
		return testPacketResult{err:err}
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
			return testPacketResult{latency: 0}
		}
		// Mystery error
		return testPacketResult{err:err}
	case n == len(buffIn):
		// Unexpectedly large read
		return testPacketResult{err: fmt.Errorf("got full buffer: %d bytes", n)}
	}

	// Unpack and and validate the message
	msg, err := message.UnmarshalMessage(buffIn)
	if err != nil {
		return testPacketResult{err:err}
	}

	err = msg.Validate()
	if err != nil {
		return testPacketResult{err:err}
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
	return testPacketResult{
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

	// packetTimerFunc tells us when to send a packet on this channel.
	sendNowChan := make(chan bool)

	// sendFromNewSocket tells us what happened on this channel.
	testResultChan := make(chan testPacketResult, 1)

	// Start the timer that will tell us when to send packets
	end := time.Now().Add(maxRTT)
	go packetTimerFunc(sendNowChan, end)

	// loop sending packets. return when we get a result (reply)
	for {
		select {
		case sendNow := <-sendNowChan:
			if sendNow {
				go func() {
					select {
					case testResultChan <- sendFromNewSocket(payload, destination, end):
					default:
					}
				}()
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
