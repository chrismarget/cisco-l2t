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
	for ticks < 6 {
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

	out := sendThis{
		payload:         payload,
		destination:     &destination,
		expectReplyFrom: destination.IP,
		rttGuess:        50 * time.Millisecond,
	}

	in := communicate(out)
	if in.err != nil {
		t.Fatal(in.err)
	}
}

func TestGoAwayBoston(t *testing.T) {
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

	out := sendThis{
		payload:         payload,
		destination:     &destination,
		expectReplyFrom: nil,
		rttGuess:        100 * time.Millisecond,
	}

	result := communicate(out)
	if result.err != nil {
		t.Fatal(result.err)
	}
}

func TestCxnTypes(t *testing.T) {
	var cxnl *net.UDPConn
	var cxnd *net.UDPConn
	var err error

	local := net.UDPAddr{
		IP: net.ParseIP("127.0.0.1"),
	}

	remote := net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 2228,
	}

	cxnl, err = net.ListenUDP(UdpProtocol, &local)
	if err != nil {
		t.Fatal(err)
	}

	cxnd, err = net.DialUDP(UdpProtocol, &local, &remote)
	if err != nil {
		t.Fatal(err)
	}

	log.Println(cxnd.RemoteAddr())
	log.Println(cxnl.RemoteAddr())

	time.Sleep(time.Second)
	_ = cxnl
	_ = cxnd
}

func TestWrongWrite(t *testing.T) {
	var cxnd *net.UDPConn
	var err error

	local := net.UDPAddr{}

	remote := net.UDPAddr{
		IP:   net.ParseIP("1.1.1.1"),
		Port: 2228,
	}

	cxnd, err = net.DialUDP(UdpProtocol, &local, &remote)
	if err != nil {
		t.Fatal(err)
	}

	n, err := cxnd.WriteToUDP([]byte{}, &remote)
	if err != nil {
		if result, ok := err.(net.Error); ok {
			log.Println(result.Error())
			log.Println(result.Timeout())
			log.Println(result.Temporary())

		}

		t.Fatal(err)
	}
	log.Println("wrote ", n)

}
