package attribute

import (
	"errors"
	"fmt"
	"net"
	"runtime"
)

const (
	macStringPrefix = "MAC Address: "
)

func stringMac(a attr) (string, error) {
	pc, _, _, _ := runtime.Caller(0)
	fname := runtime.FuncForPC(pc).Name()

	if attrCategoryByType[a.attrType] != macCategory {
		msg := fmt.Sprintf("Cannot use %s on attribute with type %d.", fname, a.attrType)
		return "", errors.New(msg)
	}

	err := a.checkLen()
	if err != nil {
		return "", err
	}

	return net.HardwareAddr(a.attrData).String(), nil
}
