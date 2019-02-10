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
	attrType attrType
	attrData []byte
}

func (o vlanAttribute) Type() attrType {
	return o.attrType
}

func (o vlanAttribute) Len() int {
	return TLsize + len(o.attrData)
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

//func newVlanAttr(t attrType, p attrPayload) (Attr, error) {
//	var result Attr
//	if p.intData < minVLAN || p.intData > maxVLAN {
//		return Attr{}, errors.New("Error creating VLAN attribute: Value out of range.")
//	}
//	result.AttrType = t
//	result.AttrData = make([]byte, 2)
//	binary.BigEndian.PutUint16(result.AttrData, uint16(p.intData))
//	return result, nil
//}
