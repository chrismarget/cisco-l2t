package attribute

import (
	"net"
)

type ipv4Attribute struct {
	attrType attrType
	attrData []byte
}

func (o ipv4Attribute) Type() attrType {
	return o.attrType
}

func (o ipv4Attribute) Len() int {
	return TLsize + len(o.attrData)
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

//// stringifyIPv4 returns a string representing the the IPv4 address in
//// dotted-quad notation. This function should be called by Attr.String()
//func stringifyIPv4(a Attr) (string, error) {
//	var err error
//	err = checkAttrInCategory(a, ipv4Category)
//	if err != nil {
//		return "", err
//	}
//
//	err = a.checkLen()
//	if err != nil {
//		return "", err
//	}
//
//	return net.IP(a.AttrData).String(), nil
//}

//// newIPv4Attr returns an Attr with AttrType t and AttrData populated based on
//// input payload. Input options are:
//// - ipAddrData (first choice)
//// - stringData (second choice, causes the function to recurse with ipAddrData)
//// - intData (last choice - needs a 32-bit compatible input, turns it into an IPv4 address)
//func newIPv4Attr(t attrType, p attrPayload) (Attr, error) {
//	result := Attr{AttrType: t}
//
//	switch {
//	case len(p.ipAddrData.IP) > 0:
//		result.AttrData = p.ipAddrData.IP.To4()
//		return result, nil
//	case p.stringData != "":
//		ip, err := net.LookupIP(p.stringData)
//		if err != nil {
//			return Attr{}, err
//		}
//		p.ipAddrData = net.IPAddr{IP: ip[0]}
//		return newIPv4Attr(t, p)
//	case p.intData >= 0:
//		if p.intData > math.MaxUint32 {
//			msg := fmt.Sprintf("Cannot create %s. Input integer data out of range: %d.", attrTypeString[t], p.intData)
//			return Attr{}, errors.New(msg)
//		}
//		b := make([]byte, 4)
//		binary.BigEndian.PutUint32(b, uint32(p.intData))
//		result.AttrData = b
//		return result, nil
//	default:
//		msg := fmt.Sprintf("Cannot create %s. No appropriate data supplied.", attrTypeString[t])
//		return Attr{}, errors.New(msg)
//	}
//}

//// validateIPv4 checks the AttrType and AttrData against norms for IPv4 type
//// attributes.
//func validateIPv4(a Attr) error {
//	if attrCategoryByType[a.AttrType] != ipv4Category{
//		msg := fmt.Sprintf("Attribute type %d cannot be validated against IPv4 criteria.", a.AttrType)
//		return errors.New(msg)
//	}
//	return nil
//}
