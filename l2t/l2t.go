package l2t

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	l2tV1   = 1
	l2tPort = 2228

	requestDst = 1
	requestSrc = 2
	replyDst   = 3
	replySrc   = 4

	srcMacAttrType        = 1
	dstMacAttrType        = 2
	vlanAttrType          = 3
	devNameAttrType       = 4
	devTypeAttrType       = 5
	devIPv4AttrType       = 6
	inPortNameAttrType    = 7
	outPortNameAttrType   = 8
	inPortSpeedAttrType   = 9
	outPortSpeedAttrType  = 10
	inPortDuplexAttrType  = 11
	outPortDuplexAttrType = 12
	nbrIPv4AttrType       = 13
	srcIPv4AttrType       = 14
	replyStatusAttrType   = 15
	nbrDevIDAttrType      = 16

	autoDuplex = 0
	halfDuplex = 1
	fullDuplex = 2

	duplexCategory = attrCategory(1)
	ipv4Category   = attrCategory(2)
	macCategory    = attrCategory(3)
	speedCategory  = attrCategory(4)
	statusCategory = attrCategory(5)
	stringCategory = attrCategory(6)
	vlanCategory   = attrCategory(7)
)

type (
	mac          [6]byte
	vlan         int16
	devName      string
	devType      string
	ip4          [4]byte
	portName     string
	portSpeed    [4]byte
	portDuplex   byte
	replyStatus  byte
	nbrDevID     string
	attrType     byte
	attrCategory string
)

type l2tAttr struct {
	attrType attrType
	attrData []byte
}

type l2tMsg struct {
	l2tMsgType byte
	l2tVer     byte
	attrs      []l2tAttr
}

var (
	l2tMsgTypeString = map[int]string{
		requestDst: "L2T_REQUEST_DST",
		requestSrc: "L2T_REQUEST_SRC",
		replyDst:   "L2T_REPLY_DST",
		replySrc:   "L2T_REPLY_SRC",
	}

	l2tAttrTypeString = map[attrType]string{
		srcMacAttrType:        "L2_ATTR_SRC_MAC",        // 6 Byte MAC address
		dstMacAttrType:        "L2_ATTR_DST_MAC",        // 6 Byte MAC address
		vlanAttrType:          "L2_ATTR_VLAN",           // 2 Byte VLAN number
		devNameAttrType:       "L2_ATTR_DEV_NAME",       // Null terminated string
		devTypeAttrType:       "L2_ATTR_DEV_TYPE",       // Null terminated string
		devIPv4AttrType:       "L2_ATTR_DEV_IP",         // 4 Byte IP Address
		inPortNameAttrType:    "L2_ATTR_INPORT_NAME",    // Null terminated string
		outPortNameAttrType:   "L2_ATTR_OUTPORT_NAME",   // Null terminated string
		inPortSpeedAttrType:   "L2_ATTR_INPORT_SPEED",   // 4 Bytes
		outPortSpeedAttrType:  "L2_ATTR_OUTPORT_SPEED",  // 4 Bytes
		inPortDuplexAttrType:  "L2_ATTR_INPORT_DUPLEX",  // 1 Byte
		outPortDuplexAttrType: "L2_ATTR_OUTPORT_DUPLEX", // 1 Byte
		nbrIPv4AttrType:       "L2_ATTR_NBR_IP",         // 4 Byte IP Address
		srcIPv4AttrType:       "L2_ATTR_SRC_IP",         // 4 Byte IP Address
		replyStatusAttrType:   "L2_ATTR_REPLY_STATUS",   // 1 Byte reply status
		nbrDevIDAttrType:      "L2_ATTR_NBR_DEV_ID",     // Null terminated string

	}

	duplexString = map[portDuplex]string{
		autoDuplex: "auto",
		halfDuplex: "half",
		fullDuplex: "full",
	}

	l2tAttrStringfuncByCategory = map[attrCategory]func([]byte) (string, error){
		"duplex": stringDuplex,
		//		"ipv4":   stringIPv4,
		"mac": stringMac,
		//		"speed":  stringSpeed,
		//		"status": stringStatus,
		//		"string": stringString,
		"vlan": stringVlan,
	}

	l2tAttrCategory = map[attrType]attrCategory{
		srcMacAttrType:        macCategory,
		dstMacAttrType:        macCategory,
		vlanAttrType:          vlanCategory,
		devNameAttrType:       stringCategory,
		devTypeAttrType:       stringCategory,
		devIPv4AttrType:       ipv4Category,
		inPortNameAttrType:    stringCategory,
		outPortNameAttrType:   stringCategory,
		inPortSpeedAttrType:   speedCategory,
		outPortSpeedAttrType:  speedCategory,
		inPortDuplexAttrType:  duplexCategory,
		outPortDuplexAttrType: duplexCategory,
		nbrIPv4AttrType:       ipv4Category,
		srcIPv4AttrType:       ipv4Category,
		replyStatusAttrType:   statusCategory,
		nbrDevIDAttrType:      stringCategory,
	}

	l2tAttrLenByCategory = map[attrCategory]int{
		duplexCategory: 3,
		ipv4Category:   6,
		macCategory:    8,
		speedCategory:  8,
		statusCategory: 3,
		stringCategory: -1,
		vlanCategory:   4,
	}
)

// checklen returns an error if the attribute's payload length doesn't make
// sense based on the claimed type. A one-byte MAC address or a seven-byte IP
// address should produce an error.
func (a l2tAttr) checkLen() error {
	var ok bool
	var category attrCategory
	if category, ok = l2tAttrCategory[a.attrType]; !ok {
		return errors.New(fmt.Sprintf("Unknown l2t Attribute type %d", a.attrType))
	}

	var expectedLen int
	if expectedLen, ok = l2tAttrLenByCategory[category]; !ok {
		msg := fmt.Sprintf("Unknown expected length for type %d (%s) l2t attributes", a.attrType, l2tAttrTypeString[a.attrType])
		return errors.New(msg)
	}

	// String type attributes have no expected length so their
	// l2tAttrLenByCategory entry will have a negative number.
	if expectedLen < 0 {
		return nil
	}

	if expectedLen != len(a.attrData)+2 {
		msg := fmt.Sprintf("Attribute type %d has invalid payload length %d", a.attrType, len(a.attrData))
		return errors.New(msg)

	}
	return nil
}

// Bytes reders an l2tAttr object into wire format as a []byte.
func (a l2tAttr) Bytes() ([]byte, error) {
	err := a.checkLen()
	if err != nil {
		return []byte{}, err
	}

	var result []byte
	result = append(result, byte(a.attrType))
	result = append(result, byte(len(a.attrData)))
	result = append(result, a.attrData...)
	return result, nil
}

func stringDuplex(d []byte) (string, error) {
	in := portDuplex(d[0])
	var result string
	var ok bool
	if result, ok = duplexString[in]; !ok {
		msg := fmt.Sprintf("Bogus duplex value: %d", in)
		return "", errors.New(msg)
	}
	return result, nil
}

func stringMac(d []byte) (string, error) {
	sep := ":"
	var result []string
	for _, v := range d {
		result = append(result, fmt.Sprintf("%02v", v))
	}
	return strings.Join(result, sep), nil
}

func stringVlan(d []byte) (string, error) {
	vlan := binary.BigEndian.Uint16(d)
	if vlan == 0 || vlan >= 4096 {
		msg := fmt.Sprintf("Bogus VLAN number: %d", vlan)
		return "", errors.New(msg)
	}
	return strconv.Itoa(int(vlan)), nil
}

// String looks up the correct l2t string method, calls it, returns the result.
func (a l2tAttr) String() (string, error) {
	err := a.checkLen()
	if err != nil {
		return "", err
	}

	var ok bool
	var cat attrCategory
	if cat, ok = l2tAttrCategory[a.attrType]; !ok {
		msg := fmt.Sprintf("Unknown l2t attribute type %d", a.attrType)
		return "", errors.New(msg)
	}

	if _, ok := l2tAttrStringfuncByCategory[cat]; !ok {
		msg := fmt.Sprintf("Don't know how to string-ify '%s' style l2t attributes (type %d)", cat, a.attrType)
		return "", errors.New(msg)
	}

	result, err := l2tAttrStringfuncByCategory[cat](a.attrData)
	if err != nil {
		return "", err
	}

	return result, nil
}

func (d portDuplex) String() string {
	if val, ok := duplexString[d]; ok {
		return val
	}
	return "unknown"
}

func MakePortDuplex(in int) portDuplex {
	return portDuplex(in)
}
