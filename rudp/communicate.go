package rudp

import (
	"fmt"
	"log"
	"net"
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

type sendThis struct {
	payload         []byte
	destination     *net.UDPAddr
	expectReplyFrom net.IP
	rttGuess        time.Duration
}

// GetOutgoingIpForDestination returns a *net.IP representing the local interface
// that's best suited for talking to the passed target address
func GetOutgoingIpForDestination(t net.IP) (net.IP, error) {
	c, err := net.Dial("udp4", t.String()+":1")
	if err != nil {
		return nil, err
	}

	return c.LocalAddr().(*net.UDPAddr).IP, c.Close()
}

func receive(cxn *net.UDPConn, source net.IP, result chan receiveResult) {
	buffIn := make([]byte, inBufferSize)
	var err error
	var received int
	var respondent *net.UDPAddr

	for received == 0 {
		received, respondent, err = cxn.ReadFromUDP(buffIn)
		switch {
		case err != nil:
			result <- receiveResult{err: err}
			return
		case received >= len(buffIn): // Unexpectedly large read
			result <- receiveResult{err: fmt.Errorf("got full buffer: %d bytes", len(buffIn))}
			return
		case source != nil && !source.Equal(respondent.IP):
			// Alien reply. Ignore.
			received = 0
			continue
		default:
		}
	}
	result <- receiveResult{
		replyFrom: respondent.IP,
		replyData: buffIn[:received],
	}
}

func transmit(cxn *net.UDPConn, destination *net.UDPAddr, payload []byte) error {
	n, err := cxn.WriteToUDP(payload, destination)
	pLen := len(payload)
	switch {
	case err != nil:
		return err
	case n != pLen:
		return fmt.Errorf("short write to socket, only manged %d of %d bytes", n, pLen)
	}
	return nil
}

func communicate(out sendThis) sendResult {
	// determine the local interface IP
	ourIp, err := GetOutgoingIpForDestination(out.destination.IP)
	if err != nil {
		return sendResult{err: err}
	}

	// create the socket
	cxn, err := net.ListenUDP(UdpProtocol, &net.UDPAddr{IP: ourIp})
	if err != nil {
		return sendResult{err: err}
	}

	// todo: this close causes ICMP unreachables. some sort of delay where the socket
	//  hangs around after we don't need it anymore would be good.
	defer func() {
		err := cxn.Close()
		if err != nil {
			log.Print("error closing socket - ", err.Error())
		}
	}()

	// socket timeout stuff
	start := time.Now()
	end := start.Add(maxRTT)
	err = cxn.SetReadDeadline(end)
	if err != nil {
		return sendResult{err: err}
	}

	replyChan := make(chan receiveResult)
	go receive(cxn, out.destination.IP, replyChan)

	// first send attempt
	err = transmit(cxn, out.destination, out.payload)
	if err != nil {
		return sendResult{err: err}
	}
	attempts := 1

	// retransmit backoff timer
	bot := NewBackoffTicker(initialRTTGuess)
	defer bot.Stop()

	// keep sending until... something happens
	for {
		select {
		case <-bot.C: // send again on RTO expiration
			err = transmit(cxn, out.destination, out.payload)
			if err != nil {
				return sendResult{err: err}
			}
			attempts++
			continue
		case <-replyChan:
			return sendResult{}
		}
	}
}
