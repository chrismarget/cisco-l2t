package communicate

import "net"

const (
	UdpProtocol = "udp4"
	IPv4 = "ipv4"
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

type SrcIPv4ForTarget interface {
	Get(*net.IP) (*net.IP, error)
}

type DefaultSrcIPv4ForTarget struct{}

func (o *DefaultSrcIPv4ForTarget) Get(target *net.IP) (*net.IP, error) {
	c, err := net.Dial(IPv4, target.String())
	if err != nil {
		return nil, err
	}
	defer c.Close()

	return &c.LocalAddr().(*net.UDPAddr).IP, nil
}
