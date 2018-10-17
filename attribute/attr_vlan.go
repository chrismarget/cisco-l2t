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

func stringifyVlan(a Attr) (string, error) {
	var err error
	err = checkAttrInCategory(a, vlanCategory)
	if err != nil {
		return "", err
	}

	err = a.checkLen()
	if err != nil {
		return "", err
	}

	vlan := binary.BigEndian.Uint16(a.AttrData)
	if vlan < minVLAN || vlan > maxVLAN {
		msg := fmt.Sprintf("Error parsing VLAN number: %d", vlan)
		return "", errors.New(msg)
	}
	return strconv.Itoa(int(vlan)), nil
}

func newVlanAttr(t attrType, p attrPayload) (Attr, error) {
	var result Attr
	if p.intData < minVLAN || p.intData > maxVLAN {
		return Attr{}, errors.New("Error creating VLAN attribute: Value out of range.")
	}
	result.AttrType = t
	result.AttrData = make([]byte, 2)
	binary.BigEndian.PutUint16(result.AttrData, uint16(p.intData))
	return result, nil
}

// validateVlan checks the AttrType and AttrData against norms for VLAN type
// attributes.
func validateVlan(a Attr) error {
	if attrCategoryByType[a.AttrType] != vlanCategory{
		msg := fmt.Sprintf("Attribute type %d cannot be validated against VLAN criteria.", a.AttrType)
		return errors.New(msg)
	}

	vlan := binary.BigEndian.Uint32(a.AttrData)
	if vlan > maxVLAN || vlan < minVLAN {
		return errors.New("Error: VLAN value out of range.")
	}
	return nil
}
