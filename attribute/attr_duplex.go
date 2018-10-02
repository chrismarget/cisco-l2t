package attribute

import (
	"errors"
	"fmt"
	"runtime"
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
	duplexString = map[portDuplex]string{
		autoDuplex: "auto",
		halfDuplex: "half",
		fullDuplex: "full",
	}
)

func stringDuplex(a attr) (string, error) {
	pc, _, _, _ := runtime.Caller(0)
	fname := runtime.FuncForPC(pc).Name()

	if attrCategoryByType[a.attrType] != duplexCategory {
		msg := fmt.Sprintf("Cannot use %s on attribute with type %d.", fname, a.attrType)
		return "", errors.New(msg)
	}

	err := a.checkLen()
	if err != nil {
		return "", err
	}

	var result string
	var ok bool
	if result, ok = duplexString[portDuplex(a.attrData[0])]; !ok {
		msg := fmt.Sprintf("Error, malformed duplex attribute: Value is %d", a.attrData)
		return "", errors.New(msg)
	}
	return result, nil
}
