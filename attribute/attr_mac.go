package attribute

import (
	"fmt"
	"net"
)

type macAttribute struct {
	attrType AttrType
	attrData []byte
}

func (o macAttribute) Type() AttrType {
	return o.attrType
}

func (o macAttribute) Len() uint8 {
	return uint8(TLsize + len(o.attrData))
}

func (o macAttribute) String() string {
	address := net.HardwareAddr(o.attrData).String()

	switch o.attrType {
	case SrcMacType:
		return address
	case DstMacType:
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

func (o macAttribute) Bytes() []byte {
	return o.attrData
}

// newMacAttribute returns a new attribute from macCategory
func (o *defaultAttrBuilder) newMacAttribute() (Attribute, error) {
	var err error
	var macAddr net.HardwareAddr
	switch {
	case o.bytesHasBeenSet:
		macAddr = o.bytesPayload
	case o.stringHasBeenSet:
		macAddr, err = net.ParseMAC(o.stringPayload)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("cannot build, no attribute payload found for category %s attribute", attrCategoryString[macCategory])
	}

	a := macAttribute{
		attrType: o.attrType,
		attrData: macAddr,
	}

	err = a.Validate()
	if err != nil {
		return nil, err
	}

	return a, nil
}
