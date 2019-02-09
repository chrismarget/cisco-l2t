package attribute

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"unicode"
)

const (
	stringTerminator = byte('\x00') // strings are null-teriminated
)

type stringAttribute struct {
	attrType attrType
	attrData []byte
}

func (o stringAttribute) Type() attrType {
	return o.attrType
}

func (o stringAttribute) Len() int {
	return TLsize + len(o.attrData)
}

func (o stringAttribute) String() string {
	return string(o.attrData[:len(o.attrData)-1])
}

func (o stringAttribute) Validate() error {
	err := checkTypeLen(o, stringCategory)
	if err != nil {
		return err
	}

	// Underlength?
	if o.Len() < TLsize+len(string(stringTerminator)) {
		return fmt.Errorf("underlength string: got %d bytes (min %d)", o.Len(), TLsize+len(string(stringTerminator)))
	}

	// Overlength?
	if o.Len() > math.MaxUint8 {
		return fmt.Errorf("overlength string: got %d bytes (max %d)", o.Len(), math.MaxUint8)
	}

	// Ends with string terminator?
	if !strings.HasSuffix(string(o.attrData), string(stringTerminator)) {
		return fmt.Errorf("string missing termination character ` %#x'", string(stringTerminator))

	}

	// Printable?
	for _, v := range o.attrData[:(len(o.attrData) - 1)] {
		if v > unicode.MaxASCII || !unicode.IsPrint(rune(v)) {
			return errors.New("string is not printable.")
		}
	}

	return nil
}

//func stringifyString(a Attr) (string, error) {
//	var err error
//	err = checkAttrInCategory(a, stringCategory)
//	if err != nil {
//		return "", err
//	}
//
//	err = a.checkLen()
//	if err != nil {
//		return "", err
//	}
//
//	trimmed := bytes.Split(a.AttrData, []byte{stringTerminator})[0]
//	if len(trimmed) != len(a.AttrData)-1 {
//		return "", errors.New("Error trimming string terminator.")
//	}
//
//	for _, v := range trimmed {
//		if v > unicode.MaxASCII || !unicode.IsPrint(rune(v)) {
//			return "", errors.New("Error, string is not printable.")
//		}
//	}
//
//	return string(trimmed), nil
//}

//func newStringAttr(t attrType, p attrPayload) (Attr, error) {
//	if p.stringData == "" {
//		return Attr{}, errors.New("Error creating string attribute: Empty string.")
//	}
//	if len(p.stringData)+TLsize >= math.MaxUint8 {
//		return Attr{}, errors.New("Error creating string attribute: Over-length string.")
//	}
//	var d []byte
//	for _, r := range p.stringData {
//		if r > unicode.MaxASCII || !unicode.IsPrint(rune(r)) {
//			return Attr{}, errors.New("Error creating string attribute: Non-string characters present.")
//		}
//		d = append(d, byte(r))
//	}
//	d = append(d, 0)
//	return Attr{AttrType: t, AttrData: d}, nil
//}

//// validateString checks the AttrType and AttrData against norms for String type
//// attributes.
//func validateString(a Attr) error {
//	if attrCategoryByType[a.AttrType] != stringCategory{
//		msg := fmt.Sprintf("Attribute type %d cannot be validated against string criteria.", a.AttrType)
//		return errors.New(msg)
//	}
//
//	trimmed := bytes.Split(a.AttrData, []byte{stringTerminator})[0]
//	if len(trimmed) != len(a.AttrData)-1 {
//		return errors.New("Error validating string termination.")
//	}
//
//	for _, v := range trimmed {
//		if v > unicode.MaxASCII || !unicode.IsPrint(rune(v)) {
//			return errors.New("Error, string is not printable.")
//		}
//	}
//
//	return nil
//}
