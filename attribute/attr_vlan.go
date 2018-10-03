package attribute

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
)

func stringVlan(a attr) (string, error) {
	var err error
	err = checkAttrInCategory(a, vlanCategory)
	if err != nil {
		return "", err
	}

	err = a.checkLen()
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
