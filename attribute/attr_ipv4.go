package attribute

import (
	"errors"
	"fmt"
	"net"
	"runtime"
)

func stringIPv4(a attr) (string, error) {
	pc, _, _, _ := runtime.Caller(0)
	fname := runtime.FuncForPC(pc).Name()

	if attrCategoryByType[a.attrType] != ipv4Category {
		msg := fmt.Sprintf("Cannot use %s on attribute with type %d.", fname, a.attrType)
		return "", errors.New(msg)
	}

	err := a.checkLen()
	if err != nil {
		return "", err
	}

	return net.IP(a.attrData).String(), nil
}
