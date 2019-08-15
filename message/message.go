package message

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/chrismarget/cisco-l2t/attribute"
	"math"
)

type (
	msgType   uint8
	msgVer    uint8
	msgLen    uint16
	attrCount uint8
)

const (
	version1       = msgVer(1)
	defaultMsgType = RequestSrc
	defaultMsgVer  = version1

	RequestDst = msgType(1)
	RequestSrc = msgType(2)
	ReplyDst   = msgType(3)
	ReplySrc   = msgType(4)

	testMacString = "ffff.ffff.ffff"
	testVlan      = 1
)

var (
	headerLenByVersion = map[msgVer]msgLen{
		version1: 5,
	}

	MsgTypeToString = map[msgType]string{
		RequestDst: "L2T_REQUEST_DST",
		RequestSrc: "L2T_REQUEST_SRC",
		ReplyDst:   "L2T_REPLY_DST",
		ReplySrc:   "L2T_REPLY_SRC",
	}

	msgTypeAttributeOrder = map[msgType][]attribute.AttrType{
		RequestDst: {
			attribute.DstMacType,
			attribute.SrcMacType,
			attribute.VlanType,
			attribute.SrcIPv4Type,
			attribute.NbrDevIDType,
		},
		RequestSrc: {
			attribute.DstMacType,
			attribute.SrcMacType,
			attribute.VlanType,
			attribute.SrcIPv4Type,
			attribute.NbrDevIDType,
		},
		ReplyDst: {
			attribute.DevNameType,
			attribute.DevTypeType,
			attribute.DevIPv4Type,
			attribute.ReplyStatusType,
			attribute.SrcIPv4Type,
			attribute.NbrDevIDType,
			attribute.InPortNameType,
			attribute.InPortDuplexType,
			attribute.InPortSpeedType,
			attribute.OutPortNameType,
			attribute.OutPortDuplexType,
			attribute.OutPortSpeedType,
		},
	}

	msgTypeRequiredAttrs = map[msgType][]attribute.AttrType{
		RequestDst: {
			attribute.DstMacType,
			attribute.SrcMacType,
			attribute.VlanType,
			attribute.SrcIPv4Type,
		},
		RequestSrc: {
			attribute.DstMacType,
			attribute.SrcMacType,
			attribute.VlanType,
			attribute.SrcIPv4Type,
		},
	}
)

// Msg represents an L2T message
type Msg interface {
	// Type returns the message type. This is the first byte on the wire.
	Type() msgType

	// Ver returns the message protocol version. This is the
	// second byte on the wire.
	Ver() msgVer

	// Len returns the message overall length: header plus sum of attribute
	// lengths. This is the third/fourth bytes on the wire
	Len() msgLen

	// AttrCount returns the count of attributes in the message. This is
	// the fifth byte on the wire.
	AttrCount() attrCount

	// Attributes returns a slice of attributes belonging to the message.
	Attributes() map[attribute.AttrType]attribute.Attribute

	// Validate checks the message for problems.
	Validate() error

	// NeedsSrcIp indicates whether this message requires an
	// attribute.SrcIPv4Type to be added before it can be sent.
	NeedsSrcIp() bool

	//	AddAttr(attribute.Attribute) attrCount
	//	DelAttr(attrCount) error

	//// SrcIpForTarget allows the caller to specify a function which picks
	//// the Type 14 (L2_ATTR_SRC_IP) payload (our IP address) when sending
	//// a message. Default behavior loads this value using egress interface
	//// address if the Type 14 attribute is omitted. There's probably no
	//// reason to call this function.
	//SrcIpForTarget(*net.IP) (*net.IP, error)

	// Marshal returns the message formatted for transmission onto the wire.
	// Extra attributes beyond those already built into the message may be
	// included when calling Marshal().
	//
	// Support for these last-minute attributes stems from the requirement to
	// include our local IP address in the L2T payload (attribute 14). When
	// Marshal()ing a message to several different targets we might need to
	// source traffic from different local IP interfaces, so this lets us tack
	// the source attribute on at the last moment.
	Marshal([]attribute.Attribute) []byte
}

type defaultMsg struct {
	msgType       msgType
	msgVer        msgVer
	attrs         map[attribute.AttrType]attribute.Attribute
	srcIpIncluded bool
	//srcIpFunc func(*net.IP) (*net.IP, error)
}

func (o *defaultMsg) Type() msgType {
	return o.msgType
}

func (o *defaultMsg) Ver() msgVer {
	return o.msgVer
}

func (o *defaultMsg) Len() msgLen {
	l := headerLenByVersion[o.msgVer]
	for _, a := range o.attrs {
		l += msgLen(a.Len())
	}
	return l
}

func (o *defaultMsg) AttrCount() attrCount {
	return attrCount(len(o.attrs))
}

func (o *defaultMsg) Attributes() map[attribute.AttrType]attribute.Attribute {
	return o.attrs
}

func (o *defaultMsg) Validate() error {
	// undersize check
	if o.Len() < headerLenByVersion[version1] {
		return fmt.Errorf("undersize message has %d bytes (min %d)", o.Len(), headerLenByVersion[version1])
	}

	// oversize check
	if o.Len() > math.MaxUint16 {
		return fmt.Errorf("oversize message has %d bytes (max %d)", o.Len(), math.MaxUint16)
	}

	// Look for duplicates, add up the length
	observedLen := headerLenByVersion[o.msgVer]
	foundAttrs := make(map[attribute.AttrType]bool)
	for _, a := range o.attrs {
		observedLen += msgLen(a.Len())
		t := a.Type()
		if _, ok := foundAttrs[a.Type()]; ok {
			return fmt.Errorf("attribute type %d (%s) repeats in message", t, attribute.AttrTypeString[t])
		}
		foundAttrs[a.Type()] = true
	}

	// length sanity check
	queriedLen := o.Len()
	if observedLen != queriedLen {
		return fmt.Errorf("wire format byte length should be %d, got %d", observedLen, queriedLen)
	}

	// attribute count sanity check
	observedAttrCount := attrCount(len(o.attrs))
	queriedAttrCount := o.AttrCount()
	if observedAttrCount != queriedAttrCount {
		return fmt.Errorf("found %d attributes, object claims to have %d", observedAttrCount, queriedAttrCount)
	}

	return nil
}

func (o *defaultMsg) NeedsSrcIp() bool {
	return !o.srcIpIncluded
}

//func (o *defaultMsg) SrcIpForTarget(t *net.IP) (*net.IP, error) {
//	return o.srcIpFunc(t)
//}

func (o *defaultMsg) Marshal(extraAttrs []attribute.Attribute) []byte {

	// extract attributes from message
	var unorderedAttrs []attribute.Attribute
	for _, a := range o.attrs {
		unorderedAttrs = append(unorderedAttrs, a)
	}

	// append extra attributes and sort
	orderedAttrs := orderAttributes(append(unorderedAttrs, extraAttrs...), o.msgType)

	attributeLen := 0
	for _, a := range orderedAttrs {
		attributeLen += int(a.Len())
	}

	// build up the 5 byte header
	marshaledLen := uint16(headerLenByVersion[version1])
	marshaledLen += uint16(attributeLen)
	lenBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(lenBytes, marshaledLen)
	var outBytes bytes.Buffer
	outBytes.Write([]byte{
		byte(o.Type()),
		byte(o.Ver()),
	})
	outBytes.Write(lenBytes)
	outBytes.Write([]byte{
		byte(len(orderedAttrs)),
	})

	for _, a := range orderedAttrs {
		aBytes := attribute.MarshalAttribute(a)
		outBytes.Write(aBytes)
	}

	return outBytes.Bytes()
}

// MsgBuilder represents an L2T message builder
type MsgBuilder interface {
	// SetType sets the message type. Only 4 types are known to exist,
	// of those, only the queries are likely relevant to this method
	// because I don't think we'll be sending replies...
	//
	// Default value is 1 (L2T_REQUEST_DST)
	SetType(msgType) MsgBuilder

	// SetVer sets the message version. Only one version is known to
	// exist, so version1 is the default.
	SetVer(msgVer) MsgBuilder

	// SetAttr adds attributes to the message's []attribute.Attribute.
	// Attribute order matters on the wire, but not within this slice.
	SetAttr(attribute.Attribute) MsgBuilder

	//// SetSrcIpFunc sets the function that will be called to calculate
	//// the attribute SrcIPv4Type (14) payload if one is required but
	//// not otherwise specified.
	//SetSrcIpFunc(func(*net.IP) (*net.IP, error)) MsgBuilder

	// Build returns a message.Msg object with the specified type,
	// version and attributes.
	Build() Msg
}

type defaultMsgBuilder struct {
	msgType msgType
	msgVer  msgVer
	attrs   map[attribute.AttrType]attribute.Attribute
	//	srcIpFunc func(*net.IP) (*net.IP, error)
}

//func defaultSrcIpFunc(target *net.IP) (*net.IP, error) {
//	c, err := net.Dial("udp4", target.String()+":1")
//	if err != nil {
//		return nil, err
//	}
//	defer c.Close()
//
//	return &c.LocalAddr().(*net.UDPAddr).IP, nil
//}

func NewMsgBuilder() MsgBuilder {
	return &defaultMsgBuilder{
		msgType: defaultMsgType,
		msgVer:  defaultMsgVer,
		//srcIpFunc: defaultSrcIpFunc,
		attrs: make(map[attribute.AttrType]attribute.Attribute),
	}
}

func (o *defaultMsgBuilder) SetType(t msgType) MsgBuilder {
	o.msgType = t
	return o
}

func (o *defaultMsgBuilder) SetVer(v msgVer) MsgBuilder {
	o.msgVer = v
	return o
}

func (o *defaultMsgBuilder) SetAttr(a attribute.Attribute) MsgBuilder {
	o.attrs[a.Type()] = a
	return o
}

//func (o *defaultMsgBuilder) SetSrcIpFunc(f func(*net.IP) (*net.IP, error)) MsgBuilder {
//	o.srcIpFunc = f
//	return o
//}

func (o *defaultMsgBuilder) Build() Msg {
	srcIpIncluded := false
	if _, exists := o.attrs[attribute.SrcIPv4Type]; exists {
		srcIpIncluded = true
	}
	m := &defaultMsg{
		msgType:       o.msgType,
		msgVer:        o.msgVer,
		attrs:         o.attrs,
		srcIpIncluded: srcIpIncluded,
		//srcIpFunc: o.srcIpFunc,
	}
	return m
}

// attrTypeLocationInSlice returns the index of the first instance
// of and attribute.AttrType within a slice, or -1 if not found
func attrTypeLocationInSlice(s []attribute.AttrType, a attribute.AttrType) int {
	for k, v := range s {
		if v == a {
			return k
		}
	}
	return -1
}

// orderAttributes sorts the supplied []Attribute according to
// the order prescribed by msgTypeAttributeOrder.
func orderAttributes(msgAttributes []attribute.Attribute, msgType msgType) []attribute.Attribute {

	// make a []AttrType that represents the input []Attribute
	var inTypes []attribute.AttrType
	for _, a := range msgAttributes {
		inTypes = append(inTypes, a.Type())
	}

	// loop over the correctly ordered []AttrType. Any attributes of the
	// appropriate type get appended to msgAttributes (they'll appear twice)
	for _, aType := range msgTypeAttributeOrder[msgType] {
		loc := attrTypeLocationInSlice(inTypes, aType)
		if loc >= 0 {
			msgAttributes = append(msgAttributes, msgAttributes[loc])
		}
	}

	// loop over all of the inTypes. Any that don't appear in the correctly
	// ordered []AttrType get appended to msgAttributes. Now everything
	// appears in the list twice.
	for _, t := range inTypes {
		loc := attrTypeLocationInSlice(msgTypeAttributeOrder[msgType], t)
		if loc < 0 {
			loc = attribute.LocationOfAttributeByType(msgAttributes, t)
			msgAttributes = append(msgAttributes, msgAttributes[loc])
		}
	}

	// At this point the msgAttributes slice is 2x its original length.
	// It begins with original data, then has required elements in order,
	// finishes with unordered elements. Cut it in half.
	targetLen := len(msgAttributes) >> 1
	msgAttributes = msgAttributes[targetLen:]

	return msgAttributes
}

// UnmarshalMessage takes a byte slice, returns a message after having
func UnmarshalMessage(b []byte) (Msg, error) {
	msg, err := UnmarshalMessageUnsafe(b)
	if err != nil {
		return nil, err
	}

	// validate the message header
	err = msg.Validate()
	if err != nil {
		return nil, err
	}

	// validate each attribute
	for _, att := range msg.Attributes() {
		err := att.Validate()
		if err != nil {
			return msg, err
		}
	}

	return msg, nil
}

func UnmarshalMessageUnsafe(b []byte) (Msg, error) {
	if len(b) < int(headerLenByVersion[version1]) {
		return nil, fmt.Errorf("cannot unmarshal message got only %d bytes", len(b))
	}

	t := msgType(b[0])
	v := msgVer(b[1])
	l := msgLen(binary.BigEndian.Uint16(b[2:4]))
	c := attrCount(b[4])

	attrs := make(map[attribute.AttrType]attribute.Attribute)

	p := int(headerLenByVersion[version1])
	for p < int(l) {
		remaining := int(l) - p
		if remaining < attribute.MinAttrLen {
			return nil, fmt.Errorf("at byte %d, not enough data remaining (%d btytes)to extract another attribute", p, attribute.MinAttrLen)
		}

		nextAttrLen := int(b[p+1])
		if remaining < nextAttrLen {
			return nil, fmt.Errorf("at byte %d, not enough data remaining to extract a %d byte attribute", p, nextAttrLen)
		}

		a, err := attribute.UnmarshalAttribute(b[p : p+nextAttrLen])
		if err != nil {
			return nil, err
		}

		attrs[a.Type()] = a
		p += nextAttrLen
	}

	if int(c) != len(attrs) {
		return nil, fmt.Errorf("header claimed %d attributes, got %d", c, len(attrs))
	}

	return &defaultMsg{
		msgType: t,
		msgVer:  v,
		attrs:   attrs,
	}, nil
}

// ListMissingAttributes takes a L2T message type and a map of attributes.
// It returns a list of attribute types that are required for this sort of
// message, but are missing from the supplied map.
func ListMissingAttributes(t msgType, a map[attribute.AttrType]attribute.Attribute) []attribute.AttrType {
	var missing []attribute.AttrType
	if required, ok := msgTypeRequiredAttrs[t]; ok {
		for _, r := range required {
			if _, ok = a[r]; !ok {
				missing = append(missing, r)
			}
		}
	}
	return missing
}

// TestMsg returns a pre-built message useful for probing for a switch
func TestMsg() (Msg, error) {
	var a attribute.Attribute
	var err error

	builder := NewMsgBuilder()
	a, err = attribute.NewAttrBuilder().SetType(attribute.SrcMacType).SetString(testMacString).Build()
	if err != nil {
		return nil, err
	}
	builder.SetAttr(a)

	a, err = attribute.NewAttrBuilder().SetType(attribute.DstMacType).SetString(testMacString).Build()
	if err != nil {
		return nil, err
	}
	builder.SetAttr(a)

	a, err = attribute.NewAttrBuilder().SetType(attribute.VlanType).SetInt(uint32(attribute.DefaultVlan)).Build()
	if err != nil {
		return nil, err
	}
	builder.SetAttr(a)

	return builder.Build(), nil
}
