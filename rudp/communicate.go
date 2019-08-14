package rudp

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

const (
	inBufferSize    = 65536
	initialRTTGuess = 100 * time.Millisecond
	maxRTT          = 2500 * time.Millisecond
	UdpProtocol     = "udp4"
)

type sendResult struct {
	err       error
	rtt       time.Duration
	sentTo    net.IP // the address we tried talking to
	sentFrom  net.IP // our IP address
	replyFrom net.IP // the address they replied from
	replyData []byte
}

type receiveResult struct {
	err error
	//	rtt       time.Duration
	replyFrom net.IP // the address they replied from
	replyData []byte
}

// timedOut returns a boolean indicating whether the receiveResult has an error
// of the Timeout variety
func (o receiveResult) timedOut() bool {
	if o.err != nil {
		if result, ok := o.err.(net.Error); ok && result.Timeout() {
			return true
		}

	}
	return false
}

type sendThis struct {
	payload         []byte
	destination     *net.UDPAddr
	expectReplyFrom net.IP
	rttGuess        time.Duration
}

// GetOutgoingIpForDestination returns a net.IP representing the local interface
// that's best suited for talking to the passed target address
func GetOutgoingIpForDestination(t net.IP) (net.IP, error) {
	c, err := net.Dial("udp4", t.String()+":1")
	if err != nil {
		return nil, err
	}

	return c.LocalAddr().(*net.UDPAddr).IP, c.Close()
}

// getRemote returns the remote *net.UDPAddr associated with a
// connected UDP socket.
func getRemote(in net.UDPConn) (*net.UDPAddr, error) {
	if in.RemoteAddr() == nil {
		// not a connected socket - un-answerable question
		return nil, fmt.Errorf("un-connected socket doesn't have a remote address")
	}

	hostString, portString, err := net.SplitHostPort(in.RemoteAddr().String())
	if err != nil {
		return &net.UDPAddr{}, err
	}

	IP := net.ParseIP(hostString)
	if IP == nil {
		return &net.UDPAddr{}, fmt.Errorf("Could not parse host: %s", hostString)
	}

	port, err := strconv.Atoi(portString)
	if err != nil {
		return &net.UDPAddr{}, err
	}

	return &net.UDPAddr{
		IP:   IP,
		Port: port,
	}, nil
}

// receiveOneMsg loops until a "good" inbound message arrives on the socket,
// or the socket times out. It ignores alien replies (packets not from
// expectedSource) unless expectedSource is <nil>. It is guaranteed to
// write to the result channel exactly once.
func receiveOneMsg(cxn *net.UDPConn, expectedSource net.IP, result chan receiveResult) {
	buffIn := make([]byte, inBufferSize)
	var err error
	var received int
	var respondent *net.UDPAddr
	var connected bool

	// pre-load respondent info for the connected socket case, because
	// the UDPConn.Read() call doesn't give us this data.
	if cxn.RemoteAddr() != nil { // "connected" socket
		connected = true
		respondent, err = getRemote(*cxn)
		if err != nil {
			result <- receiveResult{err: err}
			return
		}
	}

	// read until we have some data to return.
	for received == 0 {
		// read from the socket using the appropriate call
		switch connected {
		case true:
			received, err = cxn.Read(buffIn)
		case false:
			received, respondent, err = cxn.ReadFromUDP(buffIn)
		}

		switch {
		case err != nil:
			result <- receiveResult{err: err}
			return
		case received >= len(buffIn): // Unexpectedly large read
			result <- receiveResult{err: fmt.Errorf("got full buffer: %d bytes", len(buffIn))}
			return
		case expectedSource != nil && !expectedSource.Equal(respondent.IP):
			// Alien reply. Ignore.
			received = 0
		}
	}

	result <- receiveResult{
		replyFrom: respondent.IP,
		replyData: buffIn[:received],
	}
}

// transmit writes bytes to the specified destination using an unconnected UDP
// socket.
func transmit(cxn *net.UDPConn, destination *net.UDPAddr, payload []byte) error {
	var n int
	var err error
	if cxn.RemoteAddr() != nil {
		// connected socket created by net.DialUDP() just call Write()
		n, err = cxn.Write(payload)
	} else {
		// non-connected socket created by net.ListenUDP()
		// include the destination when calling WriteToUDP()
		n, err = cxn.WriteToUDP(payload, destination)
	}

	pLen := len(payload)
	switch {
	case err != nil:
		return err
	case n < pLen:
		return fmt.Errorf("short write to socket, only manged %d of %d bytes", n, pLen)
	case n > pLen:
		return fmt.Errorf("long write to socket, only wanted %d bytes, wrote %d bytes", pLen, n)
	}
	return nil
}

// communicate sends a message via UDP socket, collects a reply. It retransmits
// the message as needed. The input structure's expectReplyFrom is optional.
// Close the quit channel to abort the operation. Set it to <nil> if no need
// to abort.
func communicate(out sendThis, quit chan struct{}) sendResult {
	// determine the local interface IP
	ourIp, err := GetOutgoingIpForDestination(out.destination.IP)
	if err != nil {
		return sendResult{err: err}
	}

	// create the socket
	var cxn *net.UDPConn
	switch out.destination.IP.Equal(out.expectReplyFrom) {
	case true:
		cxn, err = net.DialUDP(UdpProtocol, &net.UDPAddr{IP: ourIp}, out.destination)
		if err != nil {
			return sendResult{err: err}
		}
	case false:
		cxn, err = net.ListenUDP(UdpProtocol, &net.UDPAddr{IP: ourIp})
		if err != nil {
			return sendResult{err: err}
		}
	}

	replyChan := make(chan receiveResult)
	go receiveOneMsg(cxn, out.expectReplyFrom, replyChan)

	// todo: this close causes ICMP unreachables. some sort of delay where the socket
	//  hangs around after we don't need it anymore would be good.
	defer func() {
		err := cxn.Close()
		if err != nil {
		}
	}()

	// socket timeout stuff
	start := time.Now()
	end := start.Add(maxRTT).Add(time.Minute)
	err = cxn.SetReadDeadline(end)
	if err != nil {
		return sendResult{err: err}
	}

	// first send attempt
	err = transmit(cxn, out.destination, out.payload)
	if err != nil {
		return sendResult{err: err}
	}
	outstandingMessages := 1

	// retransmit backoff timer tells us when to re-send
	bot := NewBackoffTicker(out.rttGuess)
	defer bot.Stop()

	// keep sending until... something happens
	for {
		select {
		case <-bot.C: // send again on RTO expiration
			err := transmit(cxn, out.destination, out.payload)
			if err != nil {
				return sendResult{err: err}
			}
			outstandingMessages++
		case result := <-replyChan: // reply or timeout
			if !result.timedOut() {
				outstandingMessages--
			}
			return sendResult{
				err:       result.err,
				rtt:       time.Now().Sub(start),
				sentTo:    out.destination.IP,
				replyFrom: result.replyFrom,
				replyData: result.replyData,
			}
		case <-quit: // abort
			err := cxn.SetReadDeadline(time.Now())
			if err != nil {
				return sendResult{err: err}
			}
		}
	}
}

func closeListenerAfterNReplies(cxn *net.UDPConn, replies int) {
	for i := 0; i < replies; i++ {

	}

}
