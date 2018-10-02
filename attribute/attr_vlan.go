package attribute

import (
	"encoding/binary"
	"errors"
	"fmt"
	"runtime"
	"strconv"
)

func stringVlan(a attr) (string, error) {
	pc, _, _, _ := runtime.Caller(0)
	fname := runtime.FuncForPC(pc).Name()

	if attrCategoryByType[a.attrType] != vlanCategory {
		msg := fmt.Sprintf("Cannot use %s on attribute with type %d.", fname, a.attrType)
		return "", errors.New(msg)
	}

	err := a.checkLen()
	if err != nil {
		return "", err
	}

	vlan := binary.BigEndian.Uint16(a.attrData)
	if vlan == 0 || vlan >= 4096 {
		msg := fmt.Sprintf("Error parsing VLAN number: %d", vlan)
		return "", errors.New(msg)
	}
	return strconv.Itoa(int(vlan)), nil
}
