package target

import (
	"log"
	"net"
	"testing"
)

func TestCheckTargetIP(t *testing.T) {
	ip := net.ParseIP("10.201.12.66")
	result := checkTargetIp(ip)
	if result.err != nil {
		t.Fatal(result.err)
	}
	log.Println("reply from:",result.IP)
}