package communicate

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"time"
)

const (
	inBufferSize    = 65536
	InitialRTTGuess = 100 * time.Millisecond
	MaxRTT          = 2500 * time.Millisecond
	UdpProtocol     = "udp4"
	CiscoL2TPort    = 2228
)

type SendResult struct {
	Err       error
	Aborted   bool
	Rtt       time.Duration
	SentTo    net.IP // the address we tried talking to
	SentFrom  net.IP // our IP address
	ReplyFrom net.IP // the address they replied from
	ReplyData []byte
}

type receiveResult struct {
	err error
	//	Rtt       time.Duration
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

type SendThis struct {
	Payload         []byte
	Destination     *net.UDPAddr
	ExpectReplyFrom net.IP
	RttGuess        time.Duration
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
		return &net.UDPAddr{}, fmt.Errorf("could not parse host: %s", hostString)
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
		// include the Destination when calling WriteToUDP()
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

// Communicate sends a message via UDP socket, collects a reply. It retransmits
// the message as needed. The input structure's SendThis.ExpectReplyFrom is
// optional.
//
// If SendThis.ExpectReplyFrom is populated and matches
// SendThis.Destination.IP, then a "connected" UDP socket (which can respond to
// incoming ICMP unreachables) is used.
//
// If SendThis.ExpectReplyFrom is populated and doesn't match
// SendThis.Destination.IP, then a "non-connected" (listener) UDP socket is
// used, and replies are filtered so that only datagrams sourced by
// SendThis.ExpectReplyFrom are considered.
//
// if SendThis.ExpectReplyFrom is nil, then a "non-connected" (listener) UDP
// socket is used and incoming datagrams from any source are considered valid
// replies.
//
// Close the quit channel to abort the operation. This channel can be nil if
// no need to abort. The operation is aborted by setting the receive timeout
// to "now". This has a side-effect of returning a timeout error on abort via
// the quit channel.
func Communicate(out SendThis, quit chan struct{}) SendResult {
	// determine the local interface IP
	ourIp, err := GetOutgoingIpForDestination(out.Destination.IP)
	if err != nil {
		return SendResult{Err: err}
	}

	// create the socket
	var cxn *net.UDPConn
	switch out.Destination.IP.Equal(out.ExpectReplyFrom) {
	case true:
		cxn, err = net.DialUDP(UdpProtocol, &net.UDPAddr{IP: ourIp}, out.Destination)
		if err != nil {
			return SendResult{Err: err}
		}
	case false:
		cxn, err = net.ListenUDP(UdpProtocol, &net.UDPAddr{IP: ourIp})
		if err != nil {
			return SendResult{Err: err}
		}
	}

	replyChan := make(chan receiveResult)
	go receiveOneMsg(cxn, out.ExpectReplyFrom, replyChan)

	// socket timeout stuff
	start := time.Now()
	end := start.Add(MaxRTT)
	err = cxn.SetReadDeadline(end)
	if err != nil {
		return SendResult{Err: err}
	}

	var outstandingMsgs int
	defer func() {
		go closeListenerAfterNReplies(cxn, outstandingMsgs, end)
	}()

	// retransmit backoff timer tells us when to re-send
	bot := NewBackoffTicker(out.RttGuess)
	defer bot.Stop()

	// keep track of whether the caller aborted us via the quit
	// channel
	var aborted bool

	// keep sending until... something happens
	for {
		select {
		case <-bot.C: // send again on RTO expiration
			err := transmit(cxn, out.Destination, out.Payload)
			if err != nil {
				return SendResult{Err: err}
			}
			outstandingMsgs++
		case result := <-replyChan: // reply or timeout
			if !result.timedOut() { // are we here because of a reply?
				// decrement outstanding counter on inbound reply
				outstandingMsgs--
			}
			return SendResult{
				Aborted:   aborted,
				Err:       result.err,
				Rtt:       time.Now().Sub(start),
				SentTo:    out.Destination.IP,
				ReplyFrom: result.replyFrom,
				ReplyData: result.replyData,
			}
		case <-quit: // abort
			aborted = true
			err := cxn.SetReadDeadline(time.Now())
			if err != nil {
				// note that this return happens only if the call to
				// SetReadDeadline errored (unlikely). The return on abort
				// happens on the next loop iteration via timeout error on
				// replyChan.
				return SendResult{Err: err}
			}
		}
	}
}

// closeListenerAfterNReplies
func closeListenerAfterNReplies(cxn *net.UDPConn, pendingReplies int, end time.Time) {
	var err error
	buffIn := make([]byte, inBufferSize)

	// restore the socket deadline (may have been changed due to abort)
	err = cxn.SetReadDeadline(end)
	if err != nil {
		log.Println(err)
	}

	// collect pending replies -- we really don't care what happens here.
	for i := 0; i < pendingReplies; i++ {
		// read from the socket using the appropriate call
		switch cxn.RemoteAddr() != nil {
		case true:
			// Yes I am not handling the error.
			// The only thing that matters is running out the loop.
			cxn.Read(buffIn)
		case false:
			// Yes I am not handling the error.
			// The only thing that matters is running out the loop.
			cxn.ReadFromUDP(buffIn)
		}
	}

	err = cxn.Close()
	if err != nil {
		log.Println(err)
	}
}
