package attribute

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"unicode"
)

const (
	stringTerminator = '\x00' // strings are null-teriminated
	minStringLen     = 2      // at a minimum we'll have a single character and the terminator
)

func stringifyString(a Attr) (string, error) {
	var err error
	err = checkAttrInCategory(a, stringCategory)
	if err != nil {
		return "", err
	}

	err = a.checkLen()
	if err != nil {
		return "", err
	}

	if len(a.AttrData) < minStringLen {
		msg := fmt.Sprintf("Error rendering string, only got %d byte(s).", len(a.AttrData))
		return "", errors.New(msg)
	}

	trimmed := bytes.Split(a.AttrData, []byte{stringTerminator})[0]
	if len(trimmed) != len(a.AttrData)-1 {
		return "", errors.New("Error trimming string terminator.")
	}

	for _, v := range trimmed {
		if v > unicode.MaxASCII || !unicode.IsPrint(rune(v)) {
			return "", errors.New("Error, string is not printable.")
		}
	}

	return string(trimmed), nil
}

func newStringAttr(t attrType, p attrPayload) (Attr, error) {
	if p.stringData == "" {
		return Attr{}, errors.New("Error creating string attribute: Empty string.")
	}
	if len(p.stringData)+TLsize >= math.MaxUint8 {
		return Attr{}, errors.New("Error creating string attribute: Over-length string.")
	}
	var d []byte
	for _, r := range p.stringData {
		if r > unicode.MaxASCII || !unicode.IsPrint(rune(r)) {
			return Attr{}, errors.New("Error creating string attribute: Non-string characters present.")
		}
		d = append(d, byte(r))
	}
	d = append(d, 0)
	return Attr{AttrType: t, AttrData: d}, nil
}

// validateString checks the AttrType and AttrData against norms for String type
// attributes.
func validateString(a Attr) error {
	if attrCategoryByType[a.AttrType] != stringCategory{
		msg := fmt.Sprintf("Attribute type %d cannot be validated against string criteria.", a.AttrType)
		return errors.New(msg)
	}

	trimmed := bytes.Split(a.AttrData, []byte{stringTerminator})[0]
	if len(trimmed) != len(a.AttrData)-1 {
		return errors.New("Error validating string termination.")
	}

	for _, v := range trimmed {
		if v > unicode.MaxASCII || !unicode.IsPrint(rune(v)) {
			return errors.New("Error, string is not printable.")
		}
	}

	return nil
}
