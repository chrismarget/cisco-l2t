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

	attrCategoryString = map[attrCategory]string{
		duplexCategory:      "interface duplex",
		ipv4Category:        "IPv4 address",
		macCategory:         "MAC address",
		speedCategory:       "interface speed",
		replyStatusCategory: "reply status",
		stringCategory:      "string",
		vlanCategory:        "VLAN",
	}

	stringifyAttrFuncByCategory = map[attrCategory]func(Attr) (string, error){
		duplexCategory: stringifyDuplex,
		ipv4Category:   stringifyIPv4,
		//macCategory:         stringifyMac,
		speedCategory:       stringifySpeed,
		replyStatusCategory: stringifyReplyStatus,
		stringCategory:      stringifyString,
		//vlanCategory:        stringifyVlan,
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
		ipv4Category:   validateIPv4,
		//macCategory:         validateMac,
		speedCategory:       validateSpeed,
		replyStatusCategory: validateReplyStatus,
		stringCategory:      validateString,
		//vlanCategory:        validateVlan,
	}
)

// UnmarshalAttribute returns an Attribute of the appropriate
// kind, depending on what's in the first byte (attribute type marker)
func UnmarshalAttribute(b []byte) (Attribute, error) {
	if len(b) < MinAttrLen {
		return nil, fmt.Errorf("cannot unmarshal attribute with only %d bytes (%d byte minimum)", len(b), MinAttrLen)
	}

	t := attrType(b[0])
	switch {

	case attrCategoryByType[t] == duplexCategory:
		return &duplexAttribute{attrType: t, attrData: b[1:]}, nil
	case attrCategoryByType[t] == ipv4Category:
		return &ipv4Attribute{attrType: t, attrData: b[1:]}, nil
	case attrCategoryByType[t] == macCategory:
		return &macAttribute{attrType: t, attrData: b[1:]}, nil
	case attrCategoryByType[t] == replyStatusCategory:
		return &replyStatusAttribute{attrType: t, attrData: b[1:]}, nil
	case attrCategoryByType[t] == speedCategory:
		return &speedAttribute{attrType: t, attrData: b[1:]}, nil
	case attrCategoryByType[t] == stringCategory:
		return &stringAttribute{attrType: t, attrData: b[1:]}, nil
	case attrCategoryByType[t] == vlanCategory:
		return &vlanAttribute{attrType: t, attrData: b[1:]}, nil

	}
	return nil, nil
}

type Attribute interface {
	Type() attrType
	Len() int
	String() string
	Validate() error
}

type Attr struct {
	AttrType attrType
	AttrData []byte
}

type attrPayload struct {
	intData    int
	stringData string
	ipAddrData net.IPAddr
	hwAddrData net.HardwareAddr
}

// ParseL2tAttr takes an L2T attribute as it comes from the wire ([]byte),
// renders it into an Attr structure. The length byte is validated, but is not
// part of the returned structure (measure it if needed). The resulting data is
// not checked to see whether it makes any sense (too long mac address,
// unprintable strings, etc...)
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
		AttrType: attrType(in[0]),
		AttrData: in[2:],
	}

	err := result.Validate()
	if err != nil {
		return Attr{}, err
	}

	return result, nil
}

// Validate checks the attribute length against the expected length table.
func (a Attr) Validate() error {
	err := a.checkLen()
	if err != nil {
		return err
	}

	var cat attrCategory
	var ok bool
	if cat, ok = attrCategoryByType[a.AttrType]; !ok {
		msg := fmt.Sprintf("Validation Error: Unknown attribute type %d", a.AttrType)
		return errors.New(msg)
	}

	if _, ok := validateAttrFuncByCategory[cat]; !ok {
		msg := fmt.Sprintf("Don't know how to Validate '%s' style l2t attributes (type %d)", cat, a.AttrType)
		return errors.New(msg)
	}

	validateAttrFuncByCategory[cat](a)
	if err != nil {
		return err
	}
	return nil
}

// checkLen returns an error if the attribute's payload length doesn't make
// sense based on the claimed type. A one-byte MAC address or a seven-byte IP
// address should produce an error.
func (a Attr) checkLen() error {
	var ok bool
	var category attrCategory
	if category, ok = attrCategoryByType[a.AttrType]; !ok {
		return errors.New(fmt.Sprintf("Unknown l2t Attribute type %d", a.AttrType))
	}

	expectedLen := attrLenByCategory[category] - TLsize

	// l2t attribute length field is only a single byte. We better not have more
	// data than can be described by that byte.
	if len(a.AttrData) > math.MaxUint8-TLsize {
		msg := fmt.Sprintf("Error, attribute has impossible payload size: %d bytes.", len(a.AttrData))
		return errors.New(msg)
	}

	// String type attributes have no expected length so their
	// attrLenByCategory entry will have a negative number.
	if expectedLen < 0 {
		return nil
	}

	if len(a.AttrData) != expectedLen {
		msg := fmt.Sprintf("Error, malformed %s attribute: Value length is %d.", attrTypeString[a.AttrType], len(a.AttrData))
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
	result = append(result, byte(a.AttrType))
	result = append(result, byte(len(a.AttrData)+TLsize))
	result = append(result, a.AttrData...)
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
	if cat, ok = attrCategoryByType[a.AttrType]; !ok {
		msg := fmt.Sprintf("Unknown l2t attribute type %d", a.AttrType)
		return "", errors.New(msg)
	}

	if _, ok := stringifyAttrFuncByCategory[cat]; !ok {
		msg := fmt.Sprintf("Don't know how to string-ify '%s' style l2t attributes (type %d)", cat, a.AttrType)
		return "", errors.New(msg)
	}

	result, err := stringifyAttrFuncByCategory[cat](a)
	if err != nil {
		return "", err
	}

	return result, nil
}

// NewAttr takes an AttrType and attrPayload, renders them into an Attr
// structure. Specific requirements for the contents of the attrPayload
// depend on the supplied AttrType (use intData for VLAN, stringData for
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

	if attrCategoryByType[a.AttrType] != c {
		msg := fmt.Sprintf("Cannot use %s on attribute with type %d.", fname, a.AttrType)
		return errors.New(msg)
	}

	return nil
}

// checkTypeLen checks an attribute's Attribute.Type() and Attribute.Len()
// output against norms for the supplied category.
func checkTypeLen(a Attribute, category attrCategory) error {
	if a.Len() < MinAttrLen {
		return fmt.Errorf("attribute to small: got %d bytes, need at least %d bytes", a.Len(), MinAttrLen)
	}

	if attrCategoryByType[a.Type()] != category {
		return fmt.Errorf("expected '%s' category attribute, got '%s'", attrCategoryString[attrCategoryByType[a.Type()]], attrTypeString[a.Type()])
	}

	expectedLen := attrLenByCategory[attrCategoryByType[a.Type()]]
	// Some attribute types have variable lengths.
	// Their attrLenByCategory entry is -1 (unknown).
	// Don't try to check them against the expected size table.
	if expectedLen >= MinAttrLen {
		if a.Len() != expectedLen {
			return fmt.Errorf("%s attribute should be exactly %d bytes, got %d bytes", attrTypeString[a.Type()], expectedLen, a.Len())
		}
	}
	return nil
}
