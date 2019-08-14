package target

import (
	"github.com/chrismarget/cisco-l2t/communicate"
	"log"
	"net"
	"testing"
	"time"
)

func TestNewTargetBuilder(t *testing.T) {
	tb, err := TargetBuilder().
		AddIp(net.ParseIP("192.168.8.254")).
		Build()
	if err != nil {
		t.Fatal(err)
	}

	println(tb.String())
}

func TestCheckTargetIp(t *testing.T) {
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

func TestEstimateLatency(t *testing.T){
	latency := []time.Duration{
		1 * time.Millisecond,
		2 * time.Millisecond,
		3 * time.Millisecond,
		4 * time.Millisecond,
		5 * time.Millisecond,
		6 * time.Millisecond,
		7 * time.Millisecond,
		8 * time.Millisecond,
		9 * time.Millisecond,
		10 * time.Millisecond,
		11 * time.Millisecond,
	}
	o := defaultTarget{latency: latency}
	result := o.estimateLatency()
	if result.Round(100 * time.Microsecond) != 11500 * time.Microsecond{
		t.Fatalf("expected %d usec, got %d usec", 11500, result.Round(100 * time.Microsecond))
	}
	if len(o.latency) != 10 {
		t.Fatalf("estimatLatency should truncate the latency slice to %d items", 10)
	}
	log.Println(result)
	log.Println(len(o.latency))

	latency = []time.Duration{}
	o = defaultTarget{latency: latency}
	result = o.estimateLatency()
	log.Println(result)
}
