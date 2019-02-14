package message

import (
	"fmt"
	"github.com/chrismarget/cisco-l2t/attribute"
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

func (o defaultMsg) Type() msgType {
	return o.msgType
}

func (o defaultMsg) Ver() msgVer {
	return o.msgVer
}

func (o defaultMsg) Len() msgLen {
	if o.msgLen == 0 {
		o.msgLen = headerLenByVersion[o.msgVer]
		for _, a := range o.attrs {
			o.msgLen += msgLen(a.Len())
		}
	}
	return o.msgLen
}

func (o defaultMsg) AttrCount() attrCount {
	return attrCount(len(o.attrs))
}

func (o defaultMsg) Validate() error {
	calculated := headerLenByVersion[o.msgVer]
	for _, a := range o.attrs {
		calculated += msgLen(a.Len())
	}

	result := o.Len()

	if calculated != result {
		return fmt.Errorf("Wire length should be %d, got %d", calculated, result)
	}
	return nil
}

func (o defaultMsg) Attributes() []attribute.Attribute {
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
	m := defaultMsg{
		msgType: o.msgType,
		msgVer:  o.msgVer,
		attrs:   o.attrs,
	}
	return m, nil
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

//type Msg struct {
//	msgType msgType
//	msgVer  int
//	attrs   []attribute.Attr
//}

//var (
//	msgTypeToString = map[msgType]string{
//		requestDst: "L2T_REQUEST_DST",
//		requestSrc: "L2T_REQUEST_SRC",
//		replyDst:   "L2T_REPLY_DST",
//		replySrc:   "L2T_REPLY_SRC",
//	}
//)
//
//// bytesToAttrSlice returns a []attribute.Attr. Pass it an L2T message payload
//// (all bytes in the message except the header), it'll walk the slice, parse out
//// the attributes.
//func bytesToAttrSlice(in []byte) ([]attribute.Attr, error) {
//	var result []attribute.Attr
//	inReader := bytes.NewReader(in)
//
//	tl := make([]byte, attribute.TLsize)
//	var count int
//	var err error
//
//	for {
//		// Read the Type/Length bytes
//		count, err = inReader.Read(tl)
//		if err != nil {
//			// EOF is okay here... That just means we're done picking out TLVs.
//			if err == io.EOF {
//				break
//			}
//			return nil, err
//		}
//
//		// Maybe we were already at EOF?
//		if count == 0 {
//			break
//		}
//
//		// If we got *anything*, we'd better have filled the target object.
//		if count != len(tl) {
//			return nil, errors.New("Error parsing message: Unable to read attribute header.")
//		}
//
//		// Prepare and fill an appropriately sized structure.
//		payload := make([]byte, tl[1]-attribute.TLsize)
//		count, err = inReader.Read(payload)
//		if err != nil {
//			return nil, err
//		}
//
//		// We'd better have gotten the right amount of data.
//		if count != len(payload) {
//			return nil, errors.New("Error: Underflow reading attribute payload.")
//		}
//
//		// Create an attribute object from the collected data
//		attr, err := attribute.ParseL2tAttr(append(tl, payload...))
//		if err != nil {
//			return nil, err
//		}
//
//		result = append(result, attr)
//	}
//	return result, nil
//}
//
//func (m *msgType) Validate() bool {
//	return true
//}
//
//// ParseMsg takes a []byte (probably from the network), renders it into a Msg.
//// Message format is:
////   1 byte L2T message type
////   1 byte L2T message protocol version
////   2 bytes overall message length (bytes)
////   1 byte message attribute count
////   payload (TLV data for n attributes)
//func ParseMsg(in []byte) (Msg, error) {
//	if len(in) < v1HeaderLen {
//		return Msg{}, fmt.Errorf("error parsing L2T message, got %d bytes, header alone is %d bytes", len(in), v1HeaderLen)
//	}
//
//	claimedLen := int(binary.BigEndian.Uint16(in[2:4]))
//	if claimedLen != len(in) {
//		return Msg{}, fmt.Errorf("error parsing L2T message, got %d bytes, header claims %d bytes", len(in), claimedLen)
//	}
//
//	var result Msg
//	result.msgType = msgType(in[0])
//	result.msgVer = int(in[1])
//
//	//	attrCount := int(in[4])
//	//	tlvPayload := in[5:]
//
//	//	var ok bool
//	//	_, ok = msgTypeToString[msgType], !ok
//
//	//switch {
//	//case result.msgVer != Version1:
//	//	return Msg{}, fmt.Errorf("unknown L2T version %d", ver)
//	//case !ok:
//	//	msg := fmt.Sprintf("unknown L2T type %d", msgType)
//	//	return Msg{}, errors.New(msg)
//	//case observedLen != claimedLen:
//	//	msg := fmt.Sprintf("incorrect L2T message length (header: %d, observed: %d", observedLen, claimedLen)
//	//	return Msg{}, errors.New(msg)
//	//}
//	return Msg{}, nil
//}
