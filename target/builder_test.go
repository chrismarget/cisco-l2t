package target

import (
	"log"
	"net"
	"testing"
)

func TestCheckTargetIP(t *testing.T) {
	ip := net.ParseIP("192.168.8.254")
	result := checkTargetIp(ip)
	if result.err != nil {
		t.Fatal(result.err)
	}
	log.Println("reply from:",result.IP)
}