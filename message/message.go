package message

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/chrismarget/cisco-l2t/attribute"
	"math"
	"net"
)

type (
	msgType   uint8
	msgVer    uint8
	msgLen    uint16
	attrCount uint8
)

const (
	version1       = msgVer(1)
	defaultMsgType = RequestDst
	defaultMsgVer  = version1

	RequestDst = msgType(1)
	RequestSrc = msgType(2)
	ReplyDst   = msgType(3)
	ReplySrc   = msgType(4)
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
	Attributes() []attribute.Attribute

	// Validate checks the message for problems.
	Validate() error

	AddAttr(attribute.Attribute) attrCount
	DelAttr(attrCount) error
	SrcIpForTarget(*net.IP) (*net.IP, error)

	// Marshal returns the message formatted for transmission onto the wire.
	Marshal() []byte
}

type defaultMsg struct {
	msgType   msgType
	msgVer    msgVer
	attrs     []attribute.Attribute
	msgLen    msgLen
	srcIpFunc func(*net.IP) (*net.IP, error)
}

func (o *defaultMsg) Type() msgType {
	return o.msgType
}

func (o *defaultMsg) Ver() msgVer {
	return o.msgVer
}

// TODO why zero check, just add it up dummy
// TODO for that matter, why does msgLen exist in the structure
func (o *defaultMsg) Len() msgLen {
	if o.msgLen == 0 {
		o.msgLen = headerLenByVersion[o.msgVer]
		for _, a := range o.attrs {
			o.msgLen += msgLen(a.Len())
		}
	}
	return o.msgLen
}

func (o *defaultMsg) AttrCount() attrCount {
	return attrCount(len(o.attrs))
}

func (o *defaultMsg) Attributes() []attribute.Attribute {
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
		return fmt.Errorf("Wire format byte length should be %d, got %d", observedLen, queriedLen)
	}

	// attribute count sanity check
	observedAttrCount := attrCount(len(o.attrs))
	queriedAttrCount := o.AttrCount()
	if observedAttrCount != queriedAttrCount {
		return fmt.Errorf("Found %d attributes, object claims to have %d", observedAttrCount, queriedAttrCount)
	}

	return nil
}

func (o *defaultMsg) AddAttr(a attribute.Attribute) attrCount {
	o.attrs = append(o.attrs, a)
	o.msgLen = 0
	return o.AttrCount() - 1
}

func (o *defaultMsg) DelAttr(i attrCount) error {
	if i >= o.AttrCount() {
		return fmt.Errorf("attempt to remove item %d from %d element slice", i, o.AttrCount())
	}

	copy(o.attrs[i:], o.attrs[i+1:])
	o.attrs[len(o.attrs)-1] = nil // or the zero value of T
	o.attrs = o.attrs[:len(o.attrs)-1]
	o.msgLen = 0

	return nil
}

func (o *defaultMsg) SrcIpForTarget(t *net.IP) (*net.IP, error) {
	return o.srcIpFunc(t)
}

func (o *defaultMsg) Marshal() []byte {
	// build up the 5 byte header
	lenBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(lenBytes, uint16(o.Len()))
	var outBytes bytes.Buffer
	outBytes.Write([]byte{
		byte(o.Type()),
		byte(o.Ver()),
	})
	outBytes.Write(lenBytes)
	outBytes.Write([]byte{
		byte(o.AttrCount()),
	})

	for _, a := range orderAttributes(o.attrs, o.msgType) {
		aBytes := attribute.MarshalAttribute(a)
		outBytes.Write(aBytes)
	}

	return outBytes.Bytes()
}

// MsgBuilder represents an L2T message builder
type MsgBuilder interface {
	// SetType sets the message type. Only 4 types are known to exist,
	// of those, only the queries are likely relevant to this method
	// (unless you're writing a Cisco switch replacement)
	//
	// Default value is 1 (L2T_REQUEST_DST)
	SetType(msgType) MsgBuilder

	// SetVer sets the message version. Only one version is known to
	// exist, so version1 is the default.
	SetVer(msgVer) MsgBuilder

	// AddAttr adds attributes to the message's []attribute.Attribute.
	// Attribute order matters on the wire, but not within this slice.
	AddAttr(attribute.Attribute) MsgBuilder

	// SetSrcIpFunc sets the function that will be called to calculate
	// the attribute SrcIPv4Type (14) payload if one is required but
	// not otherwise specified.
	SetSrcIpFunc(func(*net.IP) (*net.IP, error)) MsgBuilder

	// Build returns a message.Msg object with the specified type,
	// version and attributes.
	Build() Msg
}

type defaultMsgBuilder struct {
	msgType   msgType
	msgVer    msgVer
	attrs     []attribute.Attribute
	srcIpFunc func(*net.IP) (*net.IP, error)
}

func defaultSrcIpFunc(target *net.IP) (*net.IP, error) {
	c, err := net.Dial("udp4", target.String()+":1")
	if err != nil {
		return nil, err
	}
	defer c.Close()

	return &c.LocalAddr().(*net.UDPAddr).IP, nil
}

func NewMsgBuilder() MsgBuilder {
	return &defaultMsgBuilder{
		msgType:   defaultMsgType,
		msgVer:    defaultMsgVer,
		srcIpFunc: defaultSrcIpFunc,
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

func (o *defaultMsgBuilder) AddAttr(a attribute.Attribute) MsgBuilder {
	o.attrs = append(o.attrs, a)
	return o
}

func (o *defaultMsgBuilder) SetSrcIpFunc(f func(*net.IP) (*net.IP, error)) MsgBuilder {
	o.srcIpFunc = f
	return o
}

func (o *defaultMsgBuilder) Build() Msg {
	m := &defaultMsg{
		msgType:   o.msgType,
		msgVer:    o.msgVer,
		attrs:     o.attrs,
		srcIpFunc: o.srcIpFunc,
	}
	return m
}

//// LocationOfAttributeByType returns the index of the first instance
//// of an attribute.AttrType within a slice, or -1 if not found
//func LocationOfAttributeByType(s []attribute.Attribute, aType attribute.AttrType) int {
//	for i, a := range s {
//		if a.Type() == aType {
//			return i
//		}
//	}
//	return -1
//}

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

func UnmarshalMessage(b []byte) (Msg, error) {
	if len(b) < int(headerLenByVersion[version1]) {
		return nil, fmt.Errorf("cannot unmarshal message got only %d bytes", len(b))
	}

	t := msgType(b[0])
	v := msgVer(b[1])
	l := msgLen(binary.BigEndian.Uint16(b[2:4]))
	c := attrCount(b[4])

	var attrs []attribute.Attribute

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

		attrs = append(attrs, a)
		p += nextAttrLen
	}

	if int(c) != len(attrs) {
		return nil, fmt.Errorf("header claimed %d attributes, got %d", c, len(attrs))
	}

	return &defaultMsg{
		msgType: t,
		msgVer:  v,
		msgLen:  l,
		attrs:   attrs,
	}, nil
}

func AnyMissingAttributes(t msgType, a []attribute.Attribute) []attribute.AttrType {
	var missing []attribute.AttrType
	if required, ok := msgTypeRequiredAttrs[t]; ok {
		for _, r := range required {
			if attribute.LocationOfAttributeByType(a, r) < 0 {
				missing = append(missing, r)
			}
		}
	}
	return missing
}
