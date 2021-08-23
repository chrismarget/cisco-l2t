package target

import (
	"github.com/chrismarget/cisco-l2t/communicate"
	"log"
	"net"
	"testing"
)

func TestCheckTarget(t *testing.T) {
	destination := &net.UDPAddr{
		IP:   net.ParseIP("192.168.96.2"),
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

func TestTestTargetBuilder(t *testing.T) {
	log.Println("----------------------------------")
	tta, err := TestTargetBuilder().
		Build()
	if err != nil {
		t.Fatal(err)
	}
	log.Println(tta.String())
	log.Println("----------------------------------")

	ttb, err := TestTargetBuilder().
		AddIp(net.ParseIP("2.2.2.2")).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	log.Println(ttb.String())
	log.Println("----------------------------------")

	ttc, err := TestTargetBuilder().
		AddIp(net.ParseIP("3.3.3.3")).
		AddIp(net.ParseIP("4.4.4.4")).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	log.Println(ttc.String())
	log.Println("----------------------------------")

}