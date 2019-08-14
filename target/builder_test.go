package target

import (
	"net"
	"testing"
)

func TestCheckTargetIP(t *testing.T) {
	ip := net.ParseIP("192.168.15.254")
	result := checkTargetIp(ip)
	if result.err != nil {
		t.Fatal(result.err)
	}
}