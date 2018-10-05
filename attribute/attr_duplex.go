package attribute

import (
	"errors"
	"fmt"
	"log"
	"strings"
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
	portDuplexToString = map[portDuplex]string{
		autoDuplex: "auto",
		halfDuplex: "half",
		fullDuplex: "full",
	}
)

// stringifyDuplex returns a string representing a port duplex.
// This function should be called by Attr.String()
func stringifyDuplex(a Attr) (string, error) {
	err := checkAttrInCategory(a, duplexCategory)
	if err != nil {
		return "", err
	}

	err = a.checkLen()
	if err != nil {
		return "", err
	}

	var result string
	var ok bool
	if result, ok = portDuplexToString[portDuplex(a.AttrData[0])]; !ok {
		msg := fmt.Sprintf("Error, malformed duplex attribute: Value is %d", a.AttrData)
		return "", errors.New(msg)
	}
	return result, nil
}

// stringToDuplex takes a string, converts it to a []byte for use in an
// Attr.AttrData belonging to duplexCategory
func stringToDuplex(in string) ([]byte, error) {
	for k, v := range portDuplexToString {
		if strings.ToLower(v) == strings.ToLower(in) {
			result := []byte{byte(k)}
			return result, nil
		}
	}

	msg := fmt.Sprintf("Error parsing duplex string: '%s'", in)
	return []byte{}, errors.New(msg)
}

// intToDuplex takes an integer, returns a []byte for use in an
// Attr.AttrData belonging to duplexCategory
func intToDuplex(in int) ([]byte, error) {
	for k, _ := range portDuplexToString {
		if k == portDuplex(in) {
			result := []byte{byte(k)}
			return result, nil
		}
	}

	msg := fmt.Sprintf("Error parsing duplex integer: '%d'", in)
	return []byte{}, errors.New(msg)
}

// newDuplexAttr takes an AttrType (one that belongs to duplexCategory) and an
// attrPayload, parses the payload, returns a populated Attr.
func newDuplexAttr(t attrType, p attrPayload) (Attr, error) {
	result := Attr{AttrType: t}

	switch {
	case p.stringData != "":
		b, err := stringToDuplex(p.stringData)
		if err != nil {
			return Attr{}, err
		}
		result.AttrData = b
		return result, nil
	case p.intData >= 0:
		b, err := intToDuplex(p.intData)
		if err != nil {
			return Attr{}, err
		}
		result.AttrData = b
		return result, nil
	default:
		msg := fmt.Sprintf("Cannot create %s. No appropriate data supplied.", attrTypeString[t])
		return Attr{}, errors.New(msg)
	}
}

func validateDuplex(in Attr) error {
	log.Println("validateDuplex")
	return nil
}
