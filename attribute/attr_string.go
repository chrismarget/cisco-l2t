package attribute

import (
	"bytes"
	"errors"
	"fmt"
	"unicode"
)

const (
	stringTerminator   = '\x00' // strings are null-teriminated
	minStringLen       = 2      // at a minimum we'll have a single character and the terminator
)

func stringString(a attr) (string, error) {
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
