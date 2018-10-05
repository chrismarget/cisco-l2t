package attribute

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
)

const (
	minVLAN = 1
	maxVLAN = 4094
)

func stringVlan(a attr) (string, error) {
	var err error
	err = checkAttrInCategory(a, vlanCategory)
	if err != nil {
		return "", err
	}

	err = a.checkLen()
	if err != nil {
		return "", err
	}

	vlan := binary.BigEndian.Uint16(a.attrData)
	if vlan < minVLAN || vlan > maxVLAN {
		msg := fmt.Sprintf("Error parsing VLAN number: %d", vlan)
		return "", errors.New(msg)
	}
	return strconv.Itoa(int(vlan)), nil
}

func newVlanAttr(t attrType, p attrPayload) (attr, error) {
	var result attr
	if p.intData < minVLAN || p.intData > maxVLAN {
		return attr{}, errors.New("Error creating VLAN attribute: Value out of range.")
	}
	result.attrType = t
	result.attrData = make([]byte, 2)
	binary.BigEndian.PutUint16(result.attrData, uint16(p.intData))
	return result, nil
}
