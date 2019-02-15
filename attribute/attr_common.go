package attribute

import (
	"errors"
	"fmt"
	"math"
)

type (
	AttrType byte
	attrCategory string
)

const (
	TLsize     = 2
	MinAttrLen = 3

	SrcMacType        = AttrType(1)
	DstMacType        = AttrType(2)
	VlanType          = AttrType(3)
	DevNameType       = AttrType(4)
	DevTypeType       = AttrType(5)
	DevIPv4Type       = AttrType(6)
	InPortNameType    = AttrType(7)
	OutPortNameType   = AttrType(8)
	InPortSpeedType   = AttrType(9)
	OutPortSpeedType  = AttrType(10)
	InPortDuplexType  = AttrType(11)
	OutPortDuplexType = AttrType(12)
	NbrIPv4Type       = AttrType(13)
	SrcIPv4Type       = AttrType(14)
	ReplyStatusType   = AttrType(15)
	NbrDevIDType      = AttrType(16)

	duplexCategory      = attrCategory(1)
	ipv4Category        = attrCategory(2)
	macCategory         = attrCategory(3)
	speedCategory       = attrCategory(4)
	replyStatusCategory = attrCategory(5)
	stringCategory      = attrCategory(6)
	vlanCategory        = attrCategory(7)
)

var (
	attrTypeString = map[AttrType]string{
		SrcMacType:        "L2_ATTR_SRC_MAC",        // 6 Byte MAC address
		DstMacType:        "L2_ATTR_DST_MAC",        // 6 Byte MAC address
		VlanType:          "L2_ATTR_VLAN",           // 2 Byte VLAN number
		DevNameType:       "L2_ATTR_DEV_NAME",       // Null terminated string
		DevTypeType:       "L2_ATTR_DEV_TYPE",       // Null terminated string
		DevIPv4Type:       "L2_ATTR_DEV_IP",         // 4 Byte IP Address
		InPortNameType:    "L2_ATTR_INPORT_NAME",    // Null terminated string
		OutPortNameType:   "L2_ATTR_OUTPORT_NAME",   // Null terminated string
		InPortSpeedType:   "L2_ATTR_INPORT_SPEED",   // 4 Bytes
		OutPortSpeedType:  "L2_ATTR_OUTPORT_SPEED",  // 4 Bytes
		InPortDuplexType:  "L2_ATTR_INPORT_DUPLEX",  // 1 Byte
		OutPortDuplexType: "L2_ATTR_OUTPORT_DUPLEX", // 1 Byte
		NbrIPv4Type:       "L2_ATTR_NBR_IP",         // 4 Byte IP Address
		SrcIPv4Type:       "L2_ATTR_SRC_IP",         // 4 Byte IP Address
		ReplyStatusType:   "L2_ATTR_REPLY_STATUS",   // 1 Byte reply status
		NbrDevIDType:      "L2_ATTR_NBR_DEV_ID",     // Null terminated string
	}

	attrTypePrettyString = map[AttrType]string{
		SrcMacType:        "source MAC address",
		DstMacType:        "destination MAC address",
		VlanType:          "VLAN",
		DevNameType:       "device name",
		DevTypeType:       "device type",
		DevIPv4Type:       "device IPv4 address",
		InPortNameType:    "ingress port name",
		OutPortNameType:   "egress port name",
		InPortSpeedType:   "ingress port speed",
		OutPortSpeedType:  "egress port speed",
		InPortDuplexType:  "ingress port duplex",
		OutPortDuplexType: "egress port duplex",
		NbrIPv4Type:       "neighbor IPv4 address",
		SrcIPv4Type:       "source IPv4 address",
		ReplyStatusType:   "reply status",
		NbrDevIDType:      "neighbor device ID",
	}

	attrCategoryByType = map[AttrType]attrCategory{
		SrcMacType:        macCategory,
		DstMacType:        macCategory,
		VlanType:          vlanCategory,
		DevNameType:       stringCategory,
		DevTypeType:       stringCategory,
		DevIPv4Type:       ipv4Category,
		InPortNameType:    stringCategory,
		OutPortNameType:   stringCategory,
		InPortSpeedType:   speedCategory,
		OutPortSpeedType:  speedCategory,
		InPortDuplexType:  duplexCategory,
		OutPortDuplexType: duplexCategory,
		NbrIPv4Type:       ipv4Category,
		SrcIPv4Type:       ipv4Category,
		ReplyStatusType:   replyStatusCategory,
		NbrDevIDType:      stringCategory,
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
)

// MarshalAttribute returns a []byte containing a wire
// format representation of the supplied attribute.
func MarshalAttribute(a Attribute) []byte {
	t := byte(a.Type())
	l := byte(a.Len())
	b := a.Bytes()
	return append([]byte{t, l}, b...)
}

// UnmarshalAttribute returns an Attribute of the appropriate
// kind, depending on what's in the first byte (attribute type marker)
func UnmarshalAttribute(b []byte) (Attribute, error) {
	observedLength := len(b)
	if observedLength < MinAttrLen {
		return nil, fmt.Errorf("undersize attribute, cannot unmarshal %d bytes (%d byte minimum)", observedLength, MinAttrLen)
	}

	if observedLength > math.MaxUint8 {
		return nil, fmt.Errorf("oversize attribute, cannot unmarshal %d bytes (%d byte maximum)", observedLength, math.MaxUint8)
	}

	claimedLength := int(b[1])
	if observedLength != claimedLength {
		return nil, fmt.Errorf("cannot unmarshal attribute. length field says %d bytes, got %d bytes", observedLength, claimedLength)
	}

	t := AttrType(b[0])
	switch {
	case attrCategoryByType[t] == duplexCategory:
		return &duplexAttribute{attrType: t, attrData: b[2:]}, nil
	case attrCategoryByType[t] == ipv4Category:
		return &ipv4Attribute{attrType: t, attrData: b[2:]}, nil
	case attrCategoryByType[t] == macCategory:
		return &macAttribute{attrType: t, attrData: b[2:]}, nil
	case attrCategoryByType[t] == replyStatusCategory:
		return &replyStatusAttribute{attrType: t, attrData: b[2:]}, nil
	case attrCategoryByType[t] == speedCategory:
		return &speedAttribute{attrType: t, attrData: b[2:]}, nil
	case attrCategoryByType[t] == stringCategory:
		return &stringAttribute{attrType: t, attrData: b[2:]}, nil
	case attrCategoryByType[t] == vlanCategory:
		return &vlanAttribute{attrType: t, attrData: b[2:]}, nil
	}

	return nil, fmt.Errorf("cannot umarshal attribute of unknown type %d", t)
}

// Attribute represents an Attribure field from a
// Cisco L2 Traceroute packet.
type Attribute interface {

	// Type returns the Attribute's type
	Type() AttrType

	// Len returns the attribute's length. It includes the length
	// of the payload, plus 2 more bytes to cover the Type and
	// Length bytes in the header. This is the same value that
	// appears in the length field of the attribute in wire format.
	Len() uint8

	// String returns the attribute payload in printable format.
	// It does not include any descriptive wrapper stuff, does
	// well when combined with something from attrTypePrettyString.
	String() string

	// Validate returns an error if the attribute is malformed.
	Validate() error

	// Bytes returns a []byte containing the attribute payload.
	Bytes() []byte
}

// AttrBuilder builds L2T attributes.
// Calling SetType is mandatory.
// Calling one of the other "Set" methods is required
// for most values of "AttrType"
type AttrBuilder interface {
	// SetType sets the attribute type.
	SetType(AttrType) AttrBuilder

	// SetString configures the attribute with a string value.
	//
	// Use it for attributes belonging to these categories:
	//   duplexCategory: "Half" / "Full" / "Auto"
	//   ipv4Category: "x.x.x.x"
	//   macCategory: "xx:xx:xx:xx:xx:xx"
	//   replyStatusCategory "Success" / "Source Mac address not found"
	//   speedCategory: "10Mb/s" / "1Gb/s" / "10Tb/s"
	//   stringcategory: "whatever"
	//   vlanCategory: "100"
	SetString(string) AttrBuilder

	// SetInt configures the attribute with an int value.
	//
	// Use it for attributes belonging to these categories:
	//   duplexCategory
	//   ipv4Category
	//   replyStatusCategory
	//   speedCategory
	//   vlanCategory
	SetInt(uint32) AttrBuilder

	// SetBytes configures the attribute with a byte slice.
	//
	// Use it for attributes belonging to any category.
	SetBytes([]byte) AttrBuilder

	// Build builds attribute based on the AttrType and one of
	// the payloads configured with a "Set" method.
	Build() (Attribute, error)
}

type defaultAttrBuilder struct {
	attrType         AttrType
	typeHasBeenSet   bool
	stringPayload    string
	stringHasBeenSet bool
	intPayload       uint32
	intHasBeenSet    bool
	bytesPayload     []byte
	bytesHasBeenSet  bool
}

func NewAttrBuilder() AttrBuilder {
	return &defaultAttrBuilder{}
}

func (o *defaultAttrBuilder) SetType(t AttrType) AttrBuilder {
	o.attrType = t
	o.typeHasBeenSet = true
	return o
}

func (o *defaultAttrBuilder) SetString(s string) AttrBuilder {
	o.stringPayload = s
	o.stringHasBeenSet = true
	return o
}

func (o *defaultAttrBuilder) SetInt(i uint32) AttrBuilder {
	o.intPayload = i
	o.intHasBeenSet = true
	return o
}

func (o *defaultAttrBuilder) SetBytes(b []byte) AttrBuilder {
	o.bytesPayload = b
	o.bytesHasBeenSet = true
	return o
}

func (o *defaultAttrBuilder) Build() (Attribute, error) {
	if o.typeHasBeenSet != true {
		return nil, errors.New("`Attribute.Build()' called without first having set attribute type")
	}

	switch o.attrType {
	case SrcMacType, DstMacType:
		return o.newMacAttribute()
	case VlanType:
		return o.newVlanAttribute()
	case DevNameType, DevTypeType, InPortNameType, OutPortNameType, NbrDevIDType:
		return o.newStringAttribute()
	case DevIPv4Type, NbrIPv4Type, SrcIPv4Type:
		return o.newIpv4Attribute()
	case InPortSpeedType, OutPortSpeedType:
		return o.newSpeedAttribute()
	case InPortDuplexType, OutPortDuplexType:
		return o.newDuplexAttribute()
	case ReplyStatusType:
		return o.newReplyStatusAttribute()
	}
	return nil, fmt.Errorf("cannot build, unrecognized attribute type `%d'", o.attrType)
}

// checkTypeLen checks an attribute's Attribute.Type() and Attribute.Len()
// output against norms for the supplied category.
func checkTypeLen(a Attribute, category attrCategory) error {
	// Check the supplied attribute against the supplied category
	if attrCategoryByType[a.Type()] != category {
		return fmt.Errorf("expected '%s' category attribute, got '%s'", attrCategoryString[attrCategoryByType[a.Type()]], attrTypeString[a.Type()])
	}

	// An attribute should never be less than 3 bytes (including TL header)
	if a.Len() < MinAttrLen {
		return fmt.Errorf("undersize attribute: got %d bytes, need at least %d bytes", a.Len(), MinAttrLen)
	}

	// l2t attribute length field is only a single byte. We better
	// not have more data than can be described by that byte.
	if a.Len() > math.MaxUint8 {
		msg := fmt.Sprintf("oversize attribute: got %d bytes, max %d bytes", a.Len(), math.MaxUint8)
		return errors.New(msg)
	}

	// Some attribute types have variable lengths.
	// Their attrLenByCategory entry is -1 (unknown).
	// Only length check affirmative (not -1) sizes.
	expectedLen := attrLenByCategory[attrCategoryByType[a.Type()]]
	if expectedLen >= MinAttrLen {
		if int(a.Len()) != expectedLen {
			return fmt.Errorf("%s attribute should be exactly %d bytes, got %d bytes", attrTypeString[a.Type()], expectedLen, a.Len())
		}
	}
	return nil
}
