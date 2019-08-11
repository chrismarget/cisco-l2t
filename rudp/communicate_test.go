package rudp

import (
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
		t.Fatalf("expected this to take about 2400ms, but it took %s", duration)
	}
	if duration > expectedMax {
		t.Fatalf("expected this to take about 2400ms, but it took %s", duration)
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

	sockStuff := strings.Split(listenSock.LocalAddr().String(), ":")
	log.Print(sockStuff)
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
	go receive(listenSock, ip, replyChan)

	testData := make([]byte, 25)
	_, err = rand.Read(testData)
	if err != nil {
		t.Fatal(err)
	}

	//	time.Sleep(time.Second)

	log.Println("transmitting:", testData)
	err = transmit(sendSock, &destination, testData)
	if err != nil {
		t.Fatal(err)
	}

	var result receiveResult
	select {
	case result = <-replyChan:
	}
	if result.err != nil {
		t.Fatal(result.err)
	}
	log.Println("received:", result.replyData)
}
