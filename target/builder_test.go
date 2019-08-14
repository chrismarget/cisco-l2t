package target

import (
	"github.com/chrismarget/cisco-l2t/communicate"
	"log"
	"net"
	"testing"
)

func TestCheckTargetIP(t *testing.T) {
	destination := &net.UDPAddr{
		IP: net.ParseIP("192.168.8.254") ,
		Port: communicate.CiscoL2TPort,
	}
	result := checkTarget(destination)
	if result.err != nil {
		t.Fatal(result.err)
	}
	log.Println("reply from:",result.sourceIp)
}