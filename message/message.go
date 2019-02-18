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
	Version1       = msgVer(1)
	udpProtocol    = "udp4"
	udpPort        = 2228
	defaultMsgType = RequestDst
	defaultMsgVer  = Version1
	inBufferSize   = 2048

	RequestDst = msgType(1)
	RequestSrc = msgType(2)
	ReplyDst   = msgType(3)
	ReplySrc   = msgType(4)
)

var (
	HeaderLenByVersion = map[msgVer]msgLen{
		Version1: 5,
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

	// Validate checks the message for problems.
	Validate() error

	// Attributes returns a slice of attributes belonging to the message.
	Attributes() []attribute.Attribute

	// Marshal returns the message formatted for transmission onto the wire.
	Marshal() []byte

	// Communicate sends the message to the switch specified
	// in string form, waits for a reply.
	Communicate(addr *net.UDPAddr) (Msg, *net.UDPAddr, error)
}

type defaultMsg struct {
	msgType msgType
	msgVer  msgVer
	attrs   []attribute.Attribute
	msgLen  msgLen
}

// Type does blah
func (o *defaultMsg) Type() msgType {
	return o.msgType
}

func (o *defaultMsg) Ver() msgVer {
	return o.msgVer
}

func (o *defaultMsg) Len() msgLen {
	if o.msgLen == 0 {
		o.msgLen = HeaderLenByVersion[o.msgVer]
		for _, a := range o.attrs {
			o.msgLen += msgLen(a.Len())
		}
	}
	return o.msgLen
}

func (o *defaultMsg) AttrCount() attrCount {
	return attrCount(len(o.attrs))
}

func (o *defaultMsg) Validate() error {
	// undersize check
	if o.Len() < HeaderLenByVersion[Version1] {
		return fmt.Errorf("undersize message has %d bytes (min %d)", o.Len(), HeaderLenByVersion[Version1])
	}

	// oversize check
	if o.Len() > math.MaxUint16 {
		return fmt.Errorf("oversize message has %d bytes (max %d)", o.Len(), math.MaxUint16)
	}

	// Look for duplicates, add up the length
	observedLen := HeaderLenByVersion[o.msgVer]
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

	// TODO: check for required attributes for the given message type
	//       can't really do that without experimenting to find required
	//       attribute types, of course...

	return nil
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

func (o *defaultMsg) checkForRequiredAttributes() error {
	if required, ok := msgTypeRequiredAttrs[o.msgType]; ok {
		for _, r := range required {
			if locationOfAttributeByType(o.attrs, r) < 0 {
				return fmt.Errorf("message type %d cannot be sent without a type %d attribute", o.msgType, r)
			}
		}
	}
	return nil
}

func (o *defaultMsg) Communicate(target *net.UDPAddr) (Msg, *net.UDPAddr, error) {
	if target.Port == 0 {
		target.Port = udpPort
	}

	// figure out what IP we should stamp in our outgoing
	// message if it's not already specified in there
	if locationOfAttributeByType(o.attrs, attribute.SrcIPv4Type) < 0 {
		localIP, err := getLocalIpForTarget(target)
		if err != nil {
			return nil, nil, err
		}

		a, err := attribute.NewAttrBuilder().
			SetType(attribute.SrcIPv4Type).
			SetString(localIP.String()).
			Build()
		if err != nil {
			return nil, nil, err
		}

		o.attrs = append(o.attrs, a)
	}

	err := o.checkForRequiredAttributes()
	if err != nil {
		return nil, nil, err
	}

	conn, err := net.ListenUDP(udpProtocol, &net.UDPAddr{Port: udpPort})
	if err != nil {
		return nil, nil, err
	}

	payload := o.Marshal()
	n, err := conn.WriteToUDP(payload, target)
	if err != nil {
		return nil, nil, err
	}
	if n != len(payload) {
		return nil, nil, fmt.Errorf("attemtped send of %d bytes, only managed %d", len(payload), n)
	}

	buffIn := make([]byte, inBufferSize)
	n, respondent, err := conn.ReadFromUDP(buffIn)
	if n == len(buffIn) {
		return nil, respondent, fmt.Errorf("got full buffer: %d bytes", n)
	}

	err = conn.Close()
	if err != nil {
		return nil, nil, err
	}

	reply, err := UnmarshalMessage(buffIn)
	if err != nil {
		return nil, nil, err
	}

	return reply, respondent, nil
}

func (o *defaultMsg) Attributes() []attribute.Attribute {
	return o.attrs
}

type MsgBuilder interface {
	SetType(msgType) MsgBuilder
	SetVer(msgVer) MsgBuilder
	AddAttr(attribute.Attribute) MsgBuilder
	Build() (Msg, error)
}

type defaultMsgBuilder struct {
	msgType msgType
	msgVer  msgVer
	attrs   []attribute.Attribute
}

func NewMsgBuilder() MsgBuilder {
	return &defaultMsgBuilder{
		msgType: defaultMsgType,
		msgVer:  defaultMsgVer,
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

func (o *defaultMsgBuilder) Build() (Msg, error) {
	m := &defaultMsg{
		msgType: o.msgType,
		msgVer:  o.msgVer,
		attrs:   o.attrs,
	}
	return m, nil
}

// locationOfAttributeByType returns the index of the first instance
// of an attribute.AttrType within a slice, or -1 if not found
func locationOfAttributeByType(s []attribute.Attribute, aType attribute.AttrType) int {
	for i, a := range s {
		if a.Type() == aType {
			return i
		}
	}
	return -1
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
			loc = locationOfAttributeByType(msgAttributes, t)
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
	if len(b) < int(HeaderLenByVersion[Version1]) {
		return nil, fmt.Errorf("cannot unmarshal message got only %d bytes", len(b))
	}

	t := msgType(b[0])
	v := msgVer(b[1])
	l := msgLen(binary.BigEndian.Uint16(b[2:4]))
	c := attrCount(b[4])

	var attrs []attribute.Attribute

	p := int(HeaderLenByVersion[Version1])

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

func getLocalIpForTarget(target *net.UDPAddr) (*net.IP, error) {
	c, err := net.Dial(udpProtocol, target.String())
	if err != nil {
		return nil, err
	}
	defer c.Close()

	return &c.LocalAddr().(*net.UDPAddr).IP, nil
}

// TODO: defer close of UDP connection
// TODO: reaquiredAttribures test close of UDP connection
// TODO: getLocalIPforTarget test
