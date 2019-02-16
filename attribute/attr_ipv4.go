package attribute

import (
	"encoding/binary"
	"fmt"
	"net"
)

type ipv4Attribute struct {
	attrType AttrType
	attrData []byte
}

func (o ipv4Attribute) Type() AttrType {
	return o.attrType
}

func (o ipv4Attribute) Len() uint8 {
	return uint8(TLsize + len(o.attrData))
}

func (o ipv4Attribute) String() string {
	return net.IP(o.attrData).String()
}

func (o ipv4Attribute) Validate() error {
	err := checkTypeLen(o, ipv4Category)
	if err != nil {
		return err
	}
	return nil
}

func (o ipv4Attribute) Bytes() []byte {
	return o.attrData
}

// newIpv4Attribute returns a new attribute from ipv4Category
func (o *defaultAttrBuilder) newIpv4Attribute() (Attribute, error) {
	var err error
	ipv4Bytes := make([]byte, 4)
	switch {
	case o.stringHasBeenSet:
		b := net.ParseIP(o.stringPayload)
		if b == nil {
			return nil, fmt.Errorf("cannot convert `%s' to an IPv4 address", o.stringPayload)
		}
		ipv4Bytes = b[len(b)-4:]
	case o.bytesHasBeenSet:
		if len(o.bytesPayload) != 4 {
			return nil, fmt.Errorf("attempt to configure IPv4 attribute with %d byte payload", len(o.bytesPayload))
		}
		ipv4Bytes = o.bytesPayload
	case o.intHasBeenSet:
		binary.BigEndian.PutUint32(ipv4Bytes, o.intPayload)
	default:
		return nil, fmt.Errorf("cannot build, no attribute payload found for category %s attribute", attrCategoryString[ipv4Category])
	}

	a := &ipv4Attribute{
		attrType: o.attrType,
		attrData: ipv4Bytes,
	}

	err = a.Validate()
	if err != nil {
		return nil, err
	}

	return a, nil
}
