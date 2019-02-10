package attribute

import (
	"net"
)

type macAttribute struct {
	attrType attrType
	attrData []byte
}

func (o macAttribute) Type() attrType {
	return o.attrType
}

func (o macAttribute) Len() int {
	return TLsize + len(o.attrData)
}

func (o macAttribute) String() string {
	address := net.HardwareAddr(o.attrData).String()

	switch o.attrType {
	case srcMacType:
		return address
	case dstMacType:
		return address
	}
	return ""
}

func (o macAttribute) Validate() error {
	err := checkTypeLen(o, macCategory)
	if err != nil {
		return err
	}

	return nil
}

//// newMacAttr returns an Attr with AttrType t and AttrData populated based on
//// input payload. Input options are:
//// - stringData (first choice, parses the string to a MAC address)
//// - intData (second choice - renders the int as a uint64, uses the 6 low order bytes)
//func newMacAttr(t attrType, p attrPayload) (Attr, error) {
//	result := Attr{AttrType: t}
//
//	switch {
//	case p.stringData != "":
//		hw, err := net.ParseMAC(p.stringData)
//		if err != nil {
//			return Attr{}, err
//		}
//		result.AttrData = hw
//		return result, nil
//	case p.intData >= 0:
//		if uint64(p.intData) > uint64(math.Pow(2, 48)) {
//			msg := fmt.Sprintf("Cannot create %s. Input integer data out of range: %d.", attrTypeString[t], p.intData)
//			return Attr{}, errors.New(msg)
//		}
//		b := make([]byte, 8)
//		binary.BigEndian.PutUint64(b, uint64(p.intData))
//		result.AttrData = b[2:]
//		return result, nil
//	default:
//		msg := fmt.Sprintf("Cannot create %s. No appropriate data supplied.", attrTypeString[t])
//		return Attr{}, errors.New(msg)
//	}
//}
