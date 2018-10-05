package attribute

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"net"
)

// stringifyIPv4 returns a string representing the the IPv4 address in
// dotted-quad notation. This function should be called by Attr.String()
func stringifyIPv4(a Attr) (string, error) {
	var err error
	err = checkAttrInCategory(a, ipv4Category)
	if err != nil {
		return "", err
	}

	err = a.checkLen()
	if err != nil {
		return "", err
	}

	return net.IP(a.attrData).String(), nil
}

// newIPv4Attr returns an Attr with attrType t and attrData populated based on
// input payload. Input options are:
// - ipAddrData (first choice)
// - stringData (second choice, causes the function to recurse with ipAddrData)
// - intData (last choice - needs a 32-bit compatible input, turns it into an IPv4 address)
func newIPv4Attr(t attrType, p attrPayload) (Attr, error) {
	result := Attr{attrType: t}

	switch {
	case len(p.ipAddrData.IP) > 0:
		result.attrData = p.ipAddrData.IP.To4()
		return result, nil
	case p.stringData != "":
		ip, err := net.LookupIP(p.stringData)
		if err != nil {
			return Attr{}, err
		}
		p.ipAddrData = net.IPAddr{IP: ip[0]}
		return newIPv4Attr(t, p)
	case p.intData >= 0:
		if p.intData > math.MaxUint32 {
			msg := fmt.Sprintf("Cannot create %s. Input integer data out of range: %d.", attrTypeString[t], p.intData)
			return Attr{}, errors.New(msg)
		}
		b := make([]byte, 4)
		binary.BigEndian.PutUint32(b, uint32(p.intData))
		result.attrData = b
		return result, nil
	default:
		msg := fmt.Sprintf("Cannot create %s. No appropriate data supplied.", attrTypeString[t])
		return Attr{}, errors.New(msg)
	}
}
