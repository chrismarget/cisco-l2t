package l2t

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"
	"unicode"
)

const (
	stringTerminator = '\x00' // strings are null-teriminated
	minStringLen     = 2      // at a minimum we'll have a single character and the terminator
	stringStringPrefix = "String: "
)

func stringString(a attr) (string, error) {
	pc, _, _, _ := runtime.Caller(0)
	fname := runtime.FuncForPC(pc).Name()

	if attrCategoryByType[a.attrType] != stringCategory {
		msg := fmt.Sprintf("Cannot use %s on attribute with type %d.", fname, a.attrType)
		return "", errors.New(msg)
	}

	if len(a.attrData) < minStringLen {
		msg := fmt.Sprintf("Error rendering string, only got %d byte(s).", len(a.attrData))
		return "", errors.New(msg)
	}

	trimmed := bytes.Split(a.attrData, []byte{stringTerminator})[0]
	if len(trimmed) != len(a.attrData) - 1 {
		return "", errors.New("Error, trimming string terminator.")
	}

	for _, v := range(trimmed) {
		if v > unicode.MaxASCII || !unicode.IsPrint(rune(v)) {
			return "", errors.New("Error, string is not printable.")
		}
	}

	return stringStringPrefix + string(trimmed), nil
}
