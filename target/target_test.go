package target

import (
	"log"
	"net"
	"testing"
	"time"
)

func TestNewTargetBuilder(t *testing.T) {
	tb, err := NewTarget().
		AddIp(net.IP{192,168,0,1}).
		Build()
	if err != nil {
		t.Fatal(err)
	}

	println(tb.String())
}

func TestCheckTargetIp(t *testing.T) {
	testIp := net.IP{192,168,0,254}
	responseIp, latency, err := checkTargetIp(testIp)
	if err != nil {
		t.Fatal(err)
	}
	if latency <= 0 {
		t.Fatalf("observed latency %d units of duration", latency)
	}
	log.Println("testIp:",testIp)
	log.Println("responseIp:",responseIp)
	log.Println("latency:",latency)
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
}