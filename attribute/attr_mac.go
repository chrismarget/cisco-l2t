package attribute

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"net"
)

// stringifyMac takes an Attr belonging to macCategory, string-ifys it.
func stringifyMac(a Attr) (string, error) {
	var err error
	err = checkAttrInCategory(a, macCategory)
	if err != nil {
		return "", err
	}

	err = a.checkLen()
	if err != nil {
		return "", err
	}

	return net.HardwareAddr(a.attrData).String(), nil
}

// newMacAttr returns an Attr with attrType t and attrData populated based on
// input payload. Input options are:
// - stringData (first choice, parses the string to a MAC address)
// - intData (second choice - renders the int as a uint64, uses the 6 low order bytes)
func newMacAttr(t attrType, p attrPayload) (Attr, error) {
	result := Attr{attrType: t}

	switch {
	case p.stringData != "":
		hw, err := net.ParseMAC(p.stringData)
		if err != nil {
			return Attr{}, err
		}
		result.attrData = hw
		return result, nil
	case p.intData >= 0:
		if uint64(p.intData) > uint64(math.Pow(2, 48)) {
			msg := fmt.Sprintf("Cannot create %s. Input integer data out of range: %d.", attrTypeString[t], p.intData)
			return Attr{}, errors.New(msg)
		}
		b := make([]byte, 8)
		binary.BigEndian.PutUint64(b, uint64(p.intData))
		result.attrData = b[2:]
		return result, nil
	default:
		msg := fmt.Sprintf("Cannot create %s. No appropriate data supplied.", attrTypeString[t])
		return Attr{}, errors.New(msg)
	}
	return Attr{}, nil
}
