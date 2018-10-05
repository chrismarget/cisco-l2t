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

func stringString(a Attr) (string, error) {
	var err error
	err = checkAttrInCategory(a, stringCategory)
	if err != nil {
		return "", err
	}

	err = a.checkLen()
	if err != nil {
		return "", err
	}

	if len(a.attrData) < minStringLen {
		msg := fmt.Sprintf("Error rendering string, only got %d byte(s).", len(a.attrData))
		return "", errors.New(msg)
	}

	trimmed := bytes.Split(a.attrData, []byte{stringTerminator})[0]
	if len(trimmed) != len(a.attrData)-1 {
		return "", errors.New("Error, trimming string terminator.")
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
	return Attr{attrType: t, attrData: d}, nil
}
