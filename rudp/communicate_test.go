package rudp

import (
	"bytes"
	"github.com/chrismarget/cisco-l2t/attribute"
	"github.com/chrismarget/cisco-l2t/message"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestGetOutgoingIpForDestination(t *testing.T) {
	d := net.ParseIP("127.0.0.1")
	s, err := GetOutgoingIpForDestination(d)
	if err != nil {
		t.Fatal(err)
	}
	if !d.Equal(s) {
		t.Fatalf("addresses don't match: %s and %s", d, s)
	}
}

func TestReplyTimeout(t *testing.T) {
	start := time.Now()
	bot := NewBackoffTicker(100 * time.Millisecond)
	ticks := 0
	for ticks < 7 {
		select {
		case <-bot.C:
		}
		ticks++
	}
	duration := time.Now().Sub(start)
	expectedMin := 3100 * time.Millisecond
	expectedMax := 3300 * time.Millisecond
	if duration < expectedMin {
		t.Fatalf("expected this to take about 3200ms, but it took %s", duration)
	}
	if duration > expectedMax {
		t.Fatalf("expected this to take about 3200ms, but it took %s", duration)
	}
}

func TestTransmit(t *testing.T) {
	destinationIP := net.ParseIP("192.168.254.254")

	ourIp, err := GetOutgoingIpForDestination(destinationIP)
	if err != nil {
		t.Fatal(err)
	}

	cxn, err := net.ListenUDP(UdpProtocol, &net.UDPAddr{IP: ourIp})
	if err != nil {
		t.Fatal(err)
	}

	destination := net.UDPAddr{
		IP:   destinationIP,
		Port: 2228,
	}

	testData := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	err = transmit(cxn, &destination, testData)
	if err != nil {
		t.Fatal(err)
	}

	err = cxn.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestReceive(t *testing.T) {
	ip := net.ParseIP("127.0.0.1")

	listenSock, err := net.ListenUDP(UdpProtocol, &net.UDPAddr{IP: ip})
	if err != nil {
		t.Fatal(err)
	}

	destPort, err := strconv.Atoi(strings.Split(listenSock.LocalAddr().String(), ":")[1])
	if err != nil {
		t.Fatal(err)
	}

	destination := net.UDPAddr{
		IP:   ip,
		Port: destPort,
	}

	sendSock, err := net.ListenUDP(UdpProtocol, nil)
	if err != nil {
		t.Fatal(err)
	}

	replyChan := make(chan receiveResult)
	go receiveOneMsg(listenSock, ip, replyChan)

	testData := make([]byte, 25)
	_, err = rand.Read(testData)
	if err != nil {
		t.Fatal(err)
	}

	err = transmit(sendSock, &destination, testData)
	if err != nil {
		t.Fatal(err)
	}

	result := <-replyChan
	if result.err != nil {
		t.Fatal(result.err)
	}

	if !bytes.Equal(testData, result.replyData) {
		log.Fatalf("received data doesn't match sent data")
	}

	log.Println(testData)
	log.Println(result.replyData)

}

func TestCommunicateQuit(t *testing.T) {
	start := time.Now()
	destination := net.UDPAddr{
		IP:   net.ParseIP("192.168.254.254"),
		Port: 2228,
	}

	ourIp, err := GetOutgoingIpForDestination(destination.IP)
	if err != nil {
		t.Fatal(err)
	}

	ourIpAttr, err := attribute.NewAttrBuilder().
		SetType(attribute.SrcIPv4Type).
		SetString(ourIp.String()).
		Build()
	if err != nil {
		t.Fatal(err)
	}

	payload := message.TestMsg().Marshal([]attribute.Attribute{ourIpAttr})

	out := SendThis{
		Payload:         payload,
		Destination:     &destination,
		ExpectReplyFrom: destination.IP,
		RttGuess:        initialRTTGuess,
	}

	quit := make(chan struct{})
	limit := 500 * time.Millisecond

	go func() {
		time.Sleep((limit / 10) * 9)
		close(quit)
	}()

	in := Communicate(out, quit)
	duration := time.Now().Sub(start)

	if in.Err != nil {
		// timeout errors are the kind we want here.
		if result, ok := in.Err.(net.Error); ok && result.Timeout(){
			log.Println("We expected this timeout: ",in.Err)
		} else {
			t.Fatal(in.Err)
		}
	}

	if !in.Aborted {
		t.Fatalf("return doesn't indicate abort")
	}

	if duration < limit {
		log.Println("overall execution time okay")
	} else {
		t.Fatalf("Read should have completed in about %s, took longer than %s.", ((limit / 10) * 9), limit)
	}
}

func TestGoAwayBostonDial(t *testing.T) {
	destination := net.UDPAddr{
		IP:   net.ParseIP("10.201.12.66"),
		Port: 2228,
	}

	ourIp, err := GetOutgoingIpForDestination(destination.IP)
	if err != nil {
		t.Fatal(err)
	}

	ourIpAttr, err := attribute.NewAttrBuilder().
		SetType(attribute.SrcIPv4Type).
		SetString(ourIp.String()).
		Build()
	if err != nil {
		t.Fatal(err)
	}

	payload := message.TestMsg().Marshal([]attribute.Attribute{ourIpAttr})

	out := SendThis{
		Payload:         payload,
		Destination:     &destination,
		ExpectReplyFrom: destination.IP,
		RttGuess:        initialRTTGuess,
	}

	in := Communicate(out, nil)
	if in.Err != nil {
		if !strings.Contains(in.Err.Error(), "connection refused") {
			t.Fatal(in.Err)
		}
	}
}
