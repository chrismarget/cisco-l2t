package target

import (
	"github.com/chrismarget/cisco-l2t/communicate"
	"log"
	"net"
	"testing"
)

func TestCheckTarget(t *testing.T) {
	destination := &net.UDPAddr{
		IP:   net.ParseIP("192.1568.96.2"),
		Port: communicate.CiscoL2TPort,
		Zone: "",
	}
	result := checkTarget(destination)
	if result.err != nil {
		t.Fatal(result.err)
	}
	if result.latency <= 0 {
		t.Fatalf("observed latency %d units of duration", result.latency)
	}
	log.Println("testIp:",destination.IP.String())
	log.Println("responseIp:",result.sourceIp)
	log.Println("latency:",result.latency)
}
