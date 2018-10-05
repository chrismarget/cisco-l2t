package attribute

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"net"
)

// stringMac takes an attr belonging to macCategory, string-ifys it.
func stringMac(a attr) (string, error) {
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

// newMacAttr returns an attr with attrType t and attrData populated based on
// input payload. Input options are:
// - stringData (first choice, parses the string to a MAC address)
// - intData (second choice - renders the int as a uint64, uses the 6 low order bytes)
func newMacAttr(t attrType, p attrPayload) (attr, error) {
	result := attr{attrType: t}

	switch {
	case p.stringData != "":
		hw, err := net.ParseMAC(p.stringData)
		if err != nil {
			return attr{}, err
		}
		result.attrData = hw
		return result, nil
	case p.intData >= 0:
		if uint64(p.intData) > uint64(math.Pow(2, 48)) {
			msg := fmt.Sprintf("Cannot create %s. Input integer data out of range: %d.", attrTypeString[t], p.intData)
			return attr{}, errors.New(msg)
		}
		b := make([]byte, 8)
		binary.BigEndian.PutUint64(b, uint64(p.intData))
		result.attrData = b[2:]
		return result, nil
	default:
		msg := fmt.Sprintf("Cannot create %s. No appropriate data supplied.", attrTypeString[t])
		return attr{}, errors.New(msg)
	}
	return attr{}, nil
}
