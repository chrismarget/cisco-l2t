package attribute

import (
	"errors"
	"fmt"
	"math"
	"net"
	"runtime"
)

type (
	attrType     byte
	attrCategory string
)

const (
	TLsize     = 2
	MinAttrLen = 3

	srcMacType        = attrType(1)
	dstMacType        = attrType(2)
	vlanType          = attrType(3)
	devNameType       = attrType(4)
	devTypeType       = attrType(5)
	devIPv4Type       = attrType(6)
	inPortNameType    = attrType(7)
	outPortNameType   = attrType(8)
	inPortSpeedType   = attrType(9)
	outPortSpeedType  = attrType(10)
	inPortDuplexType  = attrType(11)
	outPortDuplexType = attrType(12)
	nbrIPv4Type       = attrType(13)
	srcIPv4Type       = attrType(14)
	replyStatusType   = attrType(15)
	nbrDevIDType      = attrType(16)

	duplexCategory      = attrCategory(1)
	ipv4Category        = attrCategory(2)
	macCategory         = attrCategory(3)
	speedCategory       = attrCategory(4)
	replyStatusCategory = attrCategory(5)
	stringCategory      = attrCategory(6)
	vlanCategory        = attrCategory(7)
)

var (
	attrTypeString = map[attrType]string{
		srcMacType:        "L2_ATTR_SRC_MAC",        // 6 Byte MAC address
		dstMacType:        "L2_ATTR_DST_MAC",        // 6 Byte MAC address
		vlanType:          "L2_ATTR_VLAN",           // 2 Byte VLAN number
		devNameType:       "L2_ATTR_DEV_NAME",       // Null terminated string
		devTypeType:       "L2_ATTR_DEV_TYPE",       // Null terminated string
		devIPv4Type:       "L2_ATTR_DEV_IP",         // 4 Byte IP Address
		inPortNameType:    "L2_ATTR_INPORT_NAME",    // Null terminated string
		outPortNameType:   "L2_ATTR_OUTPORT_NAME",   // Null terminated string
		inPortSpeedType:   "L2_ATTR_INPORT_SPEED",   // 4 Bytes
		outPortSpeedType:  "L2_ATTR_OUTPORT_SPEED",  // 4 Bytes
		inPortDuplexType:  "L2_ATTR_INPORT_DUPLEX",  // 1 Byte
		outPortDuplexType: "L2_ATTR_OUTPORT_DUPLEX", // 1 Byte
		nbrIPv4Type:       "L2_ATTR_NBR_IP",         // 4 Byte IP Address
		srcIPv4Type:       "L2_ATTR_SRC_IP",         // 4 Byte IP Address
		replyStatusType:   "L2_ATTR_REPLY_STATUS",   // 1 Byte reply status
		nbrDevIDType:      "L2_ATTR_NBR_DEV_ID",     // Null terminated string

	}

	attrCategoryByType = map[attrType]attrCategory{
		srcMacType:        macCategory,
		dstMacType:        macCategory,
		vlanType:          vlanCategory,
		devNameType:       stringCategory,
		devTypeType:       stringCategory,
		devIPv4Type:       ipv4Category,
		inPortNameType:    stringCategory,
		outPortNameType:   stringCategory,
		inPortSpeedType:   speedCategory,
		outPortSpeedType:  speedCategory,
		inPortDuplexType:  duplexCategory,
		outPortDuplexType: duplexCategory,
		nbrIPv4Type:       ipv4Category,
		srcIPv4Type:       ipv4Category,
		replyStatusType:   replyStatusCategory,
		nbrDevIDType:      stringCategory,
	}

	attrLenByCategory = map[attrCategory]int{
		duplexCategory:      3,
		ipv4Category:        6,
		macCategory:         8,
		speedCategory:       6,
		replyStatusCategory: 3,
		stringCategory:      -1,
		vlanCategory:        4,
	}

	stringifyAttrFuncByCategory = map[attrCategory]func(Attr) (string, error){
		duplexCategory:      stringifyDuplex,
		ipv4Category:        stringifyIPv4,
		macCategory:         stringifyMac,
		speedCategory:       stringSpeed,
		replyStatusCategory: stringifyReplyStatus,
		stringCategory:      stringifyString,
		vlanCategory:        stringifyVlan,
	}

	newAttrFuncByCategory = map[attrCategory]func(attrType, attrPayload) (Attr, error){
		duplexCategory:      newDuplexAttr,
		ipv4Category:        newIPv4Attr,
		macCategory:         newMacAttr,
		speedCategory:       newSpeedAttr,
		replyStatusCategory: newReplyStatusAttr,
		stringCategory:      newStringAttr,
		vlanCategory:        newVlanAttr,
	}

	validateAttrFuncByCategory = map[attrCategory]func(Attr) error{
		duplexCategory: validateDuplex,
		//ipv4Category:        validateIPv4,
		//macCategory:         validateMac,
		//speedCategory:       validateSpeed,
		//replyStatusCategory: validateReplyStatus,
		//stringCategory:      validateString,
		//vlanCategory:        validateVlan,
	}
)

type Attr struct {
	attrType attrType
	attrData []byte
}

type attrPayload struct {
	intData    int
	stringData string
	ipAddrData net.IPAddr
	hwAddrData net.HardwareAddr
}

// ParseL2tAttr takes an L2T attribute ([]byte) as it comes from the wire,
// renders it into an Attr structure. Length byte is validated, but is not
// part of the structure (measure it if needed). The resulting data is not
// checked to see whether it makes any sense (too long mac address, unprintable
// strings, etc...)
func ParseL2tAttr(in []byte) (Attr, error) {
	observedLen := len(in)
	if observedLen < 2 || observedLen > 255 {
		msg := fmt.Sprintf("Error parsing l2t attribute. Length cannot be %d.", observedLen)
		return Attr{}, errors.New(msg)
	}

	claimedLen := int(in[1])
	if observedLen != claimedLen {
		msg := fmt.Sprintf("Error parsing l2t attribute. Got %d bytes, but header claims %d.", observedLen, claimedLen)
		return Attr{}, errors.New(msg)
	}

	result := Attr{
		attrType: attrType(in[0]),
		attrData: in[2:],
	}

	err := result.validate()
	if err != nil {
		return Attr{}, err
	}

	return result, nil
}

// validate checks the attribute length against the expected length table.
// todo: add some payload validation
func (a Attr) validate() error {
	err := a.checkLen()
	if err != nil {
		return err
	}

	var cat attrCategory
	var ok bool
	if cat, ok = attrCategoryByType[a.attrType]; !ok {
		msg := fmt.Sprintf("Error: Unknown attribute type %d", a.attrType)
		return errors.New(msg)
	}

	//if _, ok := validateAttrFuncByCategory[cat]; !ok {
	//	msg := fmt.Sprintf("Don't know how to validate '%s' style l2t attributes (type %d)", cat, a.attrType)
	//	return errors.New(msg)
	//}
	//
	//validateAttrFuncByCategory[cat](a)
	//if err != nil {
	//	return err
	//}
	return nil
}

// checkLen returns an error if the attribute's payload length doesn't make
// sense based on the claimed type. A one-byte MAC address or a seven-byte IP
// address should produce an error.
func (a Attr) checkLen() error {
	var ok bool
	var category attrCategory
	if category, ok = attrCategoryByType[a.attrType]; !ok {
		return errors.New(fmt.Sprintf("Unknown l2t Attribute type %d", a.attrType))
	}

	expectedLen := attrLenByCategory[category] - TLsize

	// l2t attribute length field is only a single byte. We better not have more
	// data than can be described by that byte.
	if len(a.attrData) > math.MaxUint8-TLsize {
		msg := fmt.Sprintf("Error, attribute has impossible payload size: %d bytes.", len(a.attrData))
		return errors.New(msg)
	}

	// String type attributes have no expected length so their
	// attrLenByCategory entry will have a negative number.
	if expectedLen < 0 {
		return nil
	}

	if len(a.attrData) != expectedLen {
		msg := fmt.Sprintf("Error, malformed %s attribute: Value length is %d.", attrTypeString[a.attrType], len(a.attrData))
		return errors.New(msg)
	}

	return nil
}

// Bytes renders an Attr object into wire format as a []byte.
func (a Attr) Bytes() ([]byte, error) {
	err := a.checkLen()
	if err != nil {
		return []byte{}, err
	}

	var result []byte
	result = append(result, byte(a.attrType))
	result = append(result, byte(len(a.attrData)+TLsize))
	result = append(result, a.attrData...)
	return result, nil
}

// String looks up the correct l2t string method, calls it, returns the result.
func (a Attr) String() (string, error) {
	err := a.checkLen()
	if err != nil {
		return "", err
	}

	var ok bool
	var cat attrCategory
	if cat, ok = attrCategoryByType[a.attrType]; !ok {
		msg := fmt.Sprintf("Unknown l2t attribute type %d", a.attrType)
		return "", errors.New(msg)
	}

	if _, ok := stringifyAttrFuncByCategory[cat]; !ok {
		msg := fmt.Sprintf("Don't know how to string-ify '%s' style l2t attributes (type %d)", cat, a.attrType)
		return "", errors.New(msg)
	}

	result, err := stringifyAttrFuncByCategory[cat](a)
	if err != nil {
		return "", err
	}

	return result, nil
}

// NewAttr takes an attrType and attrPayload, renders them into an Attr
// structure. Specific requirements for the contents of the attrPayload
// depend on the supplied attrType (use intData for VLAN, stringData for
// strings, etc...)
func NewAttr(t attrType, p attrPayload) (Attr, error) {
	var ok bool

	// Check that we know the category
	if _, ok = attrCategoryByType[t]; !ok {
		msg := fmt.Sprintf("Unknown l2t attribute type %d", t)
		return Attr{}, errors.New(msg)
	}

	// Check that we have a "new" function for this category
	if _, ok = newAttrFuncByCategory[attrCategoryByType[t]]; !ok {
		msg := fmt.Sprintf("Don't know how to create an attribute of type '%d'", t)
		return Attr{}, errors.New(msg)
	}

	// Call the appropriate "new" function, pass it input data
	result, err := newAttrFuncByCategory[attrCategoryByType[t]](t, p)
	if err != nil {
		return Attr{}, err
	}

	return result, nil
}

// checkAttrInCategory checks whether a particular Attr belongs to the supplied
// category.
func checkAttrInCategory(a Attr, c attrCategory) error {
	pc, _, _, _ := runtime.Caller(1)
	fname := runtime.FuncForPC(pc).Name()

	if attrCategoryByType[a.attrType] != c {
		msg := fmt.Sprintf("Cannot use %s on attribute with type %d.", fname, a.attrType)
		return errors.New(msg)
	}

	return nil
}
