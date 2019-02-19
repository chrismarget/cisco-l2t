package communicate

import (
	"fmt"
	"github.com/chrismarget/cisco-l2t/attribute"
	"github.com/chrismarget/cisco-l2t/message"
	"net"
)

const (
	udpPort      = 2228
	UdpProtocol  = "udp4"
	IPv4         = "ipv4"
	inBufferSize = 2048
)

// getLocalIpForTarget returns the IP address the local host would
// use when sending to a particular target system. We use this in
// type 14 attributes (L2_ATTR_SRC_IP) when sending L2T queries.
// The field is required, but seems to be ignored: Replies come
// back to the query originator regardless of what address is
// specified via attribute 14.
func GetLocalIpForTarget(target *net.UDPAddr) (*net.IP, error) {
	c, err := net.Dial(UdpProtocol, target.String())
	if err != nil {
		return nil, err
	}
	defer c.Close()

	return &c.LocalAddr().(*net.UDPAddr).IP, nil
}

func Communicate(outmsg message.Msg, target *net.UDPAddr) (message.Msg, *net.UDPAddr, error) {
	if target.Port == 0 {
		target.Port = udpPort
	}

	// Figure out whether SrcIPv4Type is among missing attributes for this message
	missing := message.AnyMissingAttributes(outmsg.Type(), outmsg.Attributes())
	needToAddSrcIPv4Type := false
	for _, m := range missing {
		if m == attribute.SrcIPv4Type {
			needToAddSrcIPv4Type = true
		}
	}

	var payload []byte

	// Add the SrcIPv4Type attribute if necessary
	if needToAddSrcIPv4Type {
		localIP, err := outmsg.SrcIpForTarget(&target.IP)
		if err != nil {
			return nil, nil, err
		}

		a, err := attribute.NewAttrBuilder().
			SetType(attribute.SrcIPv4Type).
			SetString(localIP.String()).
			Build()
		if err != nil {
			return nil, nil, err
		}

		// TODO some locking here while adding/removing attributes?
		i := outmsg.AddAttr(a)
		payload = outmsg.Marshal()
		err = outmsg.DelAttr(i)
		if err != nil {
			return nil, nil, err
		}
	} else {
		payload = outmsg.Marshal()
	}

	conn, err := net.ListenUDP(UdpProtocol, &net.UDPAddr{})
	if err != nil {
		return nil, nil, err
	}
	defer conn.Close()

	n, err := conn.WriteToUDP(payload, target)
	if err != nil {
		return nil, nil, err
	}
	if n != len(payload) {
		return nil, nil, fmt.Errorf("attemtped send of %d bytes, only managed %d", len(payload), n)
	}

	buffIn := make([]byte, inBufferSize)
	n, respondent, err := conn.ReadFromUDP(buffIn)
	if n == len(buffIn) {
		return nil, respondent, fmt.Errorf("got full buffer: %d bytes", n)
	}

	reply, err := message.UnmarshalMessage(buffIn)
	if err != nil {
		return nil, nil, err
	}

	return reply, respondent, nil
}
