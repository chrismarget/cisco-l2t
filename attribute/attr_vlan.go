package attribute

import (
	"encoding/binary"
	"fmt"
	"strconv"
)

const (
	minVLAN = 1
	maxVLAN = 4094
)

type vlanAttribute struct {
	attrType AttrType
	attrData []byte
}

func (o vlanAttribute) Type() AttrType {
	return o.attrType
}

func (o vlanAttribute) Len() uint8 {
	return uint8(TLsize + len(o.attrData))
}

func (o vlanAttribute) String() string {
	vlan := binary.BigEndian.Uint16(o.attrData[0:2])
	return strconv.Itoa(int(vlan))
}

func (o vlanAttribute) Validate() error {
	err := checkTypeLen(o, vlanCategory)
	if err != nil {
		return err
	}

	vlan := binary.BigEndian.Uint16(o.attrData)
	if vlan > maxVLAN || vlan < minVLAN {
		return fmt.Errorf("VLAN %d value out of range", vlan)
	}

	return nil
}

func (o vlanAttribute) Bytes() []byte {
	return o.attrData
}

// newVlanAttribute returns a new attribute from vlanCategory
func (o *defaultAttrBuilder) newVlanAttribute() (Attribute, error) {
	var err error
	vlanBytes := make([]byte, 2)
	switch {
	case o.stringHasBeenSet:
		vlan, err := strconv.Atoi(o.stringPayload)
		if err != nil {
			return nil, err
		}
		binary.BigEndian.PutUint16(vlanBytes, uint16(vlan))
	case o.intHasBeenSet:
		vlan := int(o.intPayload)
		binary.BigEndian.PutUint16(vlanBytes, uint16(vlan))
	case o.bytesHasBeenSet:
		vlanBytes = o.bytesPayload
	default:
		return nil, fmt.Errorf("cannot build, no attribute payload found for category %s attribute", attrCategoryString[vlanCategory])
	}

	a := vlanAttribute{
		attrType: o.attrType,
		attrData: vlanBytes,
	}

	err = a.Validate()
	if err != nil {
		return nil, err
	}

	return a, nil
}
