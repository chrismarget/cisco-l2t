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
	udpPort        = 2228
	defaultMsgType = requestDst
	defaultMsgVer  = version1

	requestDst = msgType(1)
	requestSrc = msgType(2)
	replyDst   = msgType(3)
	replySrc   = msgType(4)
)

var (
	headerLenByVersion = map[msgVer]msgLen{
		version1: 5,
	}

	msgTypeToString = map[msgType]string{
		requestDst: "L2T_REQUEST_DST",
		requestSrc: "L2T_REQUEST_SRC",
		replyDst:   "L2T_REPLY_DST",
		replySrc:   "L2T_REPLY_SRC",
	}
)

type Msg interface {
	Type() msgType
	Ver() msgVer
	Len() msgLen
	AttrCount() attrCount
	Validate() error
	Attributes() []attribute.Attribute
}

type defaultMsg struct {
	msgType msgType
	msgVer  msgVer
	attrs   []attribute.Attribute
	msgLen  msgLen
}

func (o *defaultMsg) Type() msgType {
	return o.msgType
}

func (o *defaultMsg) Ver() msgVer {
	return o.msgVer
}

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

func MarshalMsg(msg Msg) []byte {
	// build up the 5 byte header
	msglen := make([]byte, 2)
	binary.BigEndian.PutUint16(msglen, uint16(msg.Len()))
	var b bytes.Buffer
	b.Write([]byte{
		byte(msg.Type()),
		byte(msg.Ver()),
	})
	b.Write(msglen)
	b.Write([]byte{
		byte(msg.AttrCount()),
	})

	for _, a := range msg.Attributes() {
		aBytes := attribute.MarshalAttribute(a)
		b.Write(aBytes)
	}

	return b.Bytes()
}

func orderAttributesMsgRequestSrc(in []attribute.Attribute) (attribute.Attribute, error) {
	var out []attribute.Attribute
	correctOrder := []attribute.AttrType{
		attribute.DstMacType,
		attribute.SrcMacType,
		attribute.VlanType,
		attribute.DevIPv4Type,
		attribute.NbrDevIDType,
	}
	optionalAttribute := []attribute.AttrType{
		attribute.NbrDevIDType,
		attribute.VlanType,
	}
	_ = out
	_ = correctOrder
	_ = optionalAttribute
	return nil, nil
}
