package attribute

import (
	"fmt"
	"errors"
	"strings"
)

type (
	portDuplex byte
)

const (
	autoDuplex = portDuplex(0)
	halfDuplex = portDuplex(1)
	fullDuplex = portDuplex(2)
)

var (
	portDuplexToString = map[portDuplex]string{
		autoDuplex: "Auto",
		halfDuplex: "Half",
		fullDuplex: "Full",
	}
)

type duplexAttribute struct {
	attrType AttrType
	attrData []byte
}

func (o duplexAttribute) Type() AttrType {
	return o.attrType
}

func (o duplexAttribute) Len() uint8 {
	return uint8(TLsize + len(o.attrData))
}

func (o duplexAttribute) String() string {
	return portDuplexToString[portDuplex(o.attrData[0])]
}

func (o duplexAttribute) Validate() error {
	err := checkTypeLen(o, duplexCategory)
	if err != nil {
		return err
	}

	if _, ok := portDuplexToString[portDuplex(o.attrData[0])]; !ok {
		return fmt.Errorf("`%#x' not a valid payload for %s", o.attrData[0], AttrTypeString[o.attrType])
	}

	return nil
}

func (o duplexAttribute) Bytes() []byte {
	return o.attrData
}

// newDuplexAttribute returns a new attribute from duplexCategory
func (o *defaultAttrBuilder) newDuplexAttribute() (Attribute, error) {
	var duplexByte byte
	var success bool
	switch {
	case o.stringHasBeenSet:
		for portDuplex, duplexString := range portDuplexToString {
			if strings.ToLower(o.stringPayload) == strings.ToLower(duplexString) {
				duplexByte = byte(portDuplex)
				success = true
			}
		}
		if !success {
			return nil, fmt.Errorf("string payload `%s' unrecognized for duplex type", o.stringPayload)
		}
	case o.intHasBeenSet:
		for portDuplex, _ := range portDuplexToString {
			if uint8(o.intPayload) == uint8(portDuplex) {
				duplexByte = byte(portDuplex)
				success = true
			}
		}
		if !success {
			return nil, fmt.Errorf("int payload `%d' unrecognized for duplex type", o.intPayload)
		}
	case o.bytesHasBeenSet:
		if len(o.bytesPayload) != 1 {
			return nil, errors.New("bytes payload invalid length for creating duplex attribute")
		}
		duplexByte = o.bytesPayload[0]
	default:
		return nil, fmt.Errorf("cannot build, no attribute payload found for category %s attribute", attrCategoryString[duplexCategory])
	}

	a := &duplexAttribute{
		attrType: o.attrType,
		attrData: []byte{duplexByte},
	}

	err := a.Validate()
	if err != nil {
		return nil, err
	}

	return a, nil
}
