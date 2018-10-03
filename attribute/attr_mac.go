package attribute

import (
	"net"
)

const (
	macStringPrefix = "MAC Address: "
)

func stringMac(a attr) (string, error) {
	var err error
	err = checkAttrInCategory(a, macCategory)
	if err != nil {
		return "", err
	}

	err = a.checkLen()
	if err != nil {
		return "", err
	}

	return net.HardwareAddr(a.attrData).String(), nil
}
