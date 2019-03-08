package target

import (
	"log"
	"net"
	"testing"
)

func TestNewTargetBuilder(t *testing.T) {
	tb, err := NewTarget().
		AddIp(net.IP{192,168,0,1}).
		AddIp(net.IP{192,168,0,254}).
		AddIp(net.IP{192,168,0,252}).
		AddIp(net.IP{192,168,0,1}).
		Build()
	if err != nil {
		t.Fatal(err)
	}

	println(tb.String())
}

func TestCheckTargetIp(t *testing.T) {
	testIp := net.IP{192,168,0,254}
	responseIp, err := checkTargetIp(&testIp)
	if err != nil {
		t.Fatal(err)
	}
	log.Println(testIp)
	log.Println(responseIp)
}