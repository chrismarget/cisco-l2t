package message

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/chrismarget/cisco-l2t/attribute"
	"io"
)

type (
	msgType int
)

const (
	Version1 = 1
	//udpPort = 2228
	v1HeaderLen = 5

	requestDst = msgType(1)
	requestSrc = msgType(2)
	replyDst   = msgType(3)
	replySrc   = msgType(4)
)

type Msg struct {
	msgType msgType
	msgVer  int
	attrs   []attribute.Attr
}

var (
	msgTypeToString = map[msgType]string{
		requestDst: "L2T_REQUEST_DST",
		requestSrc: "L2T_REQUEST_SRC",
		replyDst:   "L2T_REPLY_DST",
		replySrc:   "L2T_REPLY_SRC",
	}
)

// bytesToAttrSlice returns a []attribute.Attr. Pass it an L2T message payload
// (all bytes in the message except the header), it'll walk the slice, parse out
// the attributes.
func bytesToAttrSlice(in []byte) ([]attribute.Attr, error) {
	var result []attribute.Attr
	inReader := bytes.NewReader(in)

	tl := make([]byte, attribute.TLsize)
	var count int
	var err error

	for {
		// Read the Type/Length bytes
		count, err = inReader.Read(tl)
		if err != nil {
			// EOF is okay here... That just means we're done picking out TLVs.
			if err == io.EOF {
				break
			}
			return nil, err
		}

		// Maybe we were already at EOF?
		if count == 0 {
			break
		}

		// If we got *anything*, we'd better have filled the target object.
		if count != len(tl) {
			return nil, errors.New("Error parsing message: Unable to read attribute header.")
		}

		// Prepare and fill an appropriately sized structure.
		payload := make([]byte, tl[1]-attribute.TLsize)
		count, err = inReader.Read(payload)
		if err != nil {
			return nil, err
		}

		// We'd better have gotten the right amount of data.
		if count != len(payload) {
			return nil, errors.New("Error: Underflow reading attribute payload.")
		}

		// Create an attribute object from the collected data
		attr, err := attribute.ParseL2tAttr(append(tl, payload...))
		if err != nil {
			return nil, err
		}

		result = append(result, attr)
	}
	return result, nil
}

func (m *msgType) Validate() bool {
	return true
}

// ParseMsg takes a []byte (probably from the network), renders it into a Msg.
// Message format is:
//   1 byte L2T message type
//   1 byte L2T message protocol version
//   2 bytes overall message length (bytes)
//   1 byte message attribute count
//   payload (TLV data for n attributes)
func ParseMsg(in []byte) (Msg, error) {
	if len(in) < v1HeaderLen {
		msg := fmt.Sprintf("Error parsing L2T message, only got %d bytes.", len(in))
		return Msg{}, errors.New(msg)
	}

	var result Msg
	result.msgType = msgType(in[0])
	result.msgVer = int(in[1])

	msgType := msgType(in[0])
	ver := int(in[1])
	claimedLen := int(binary.BigEndian.Uint16(in[2:4]))
	observedLen := len(in)
	attrCount := int(in[4])
	tlvPayload := in[5:]

	var ok bool
	_, ok = msgTypeToString[msgType], !ok

	switch {
	case ver != Version1:
		msg := fmt.Sprintf("Error: Unknown L2T version %d", ver)
		return Msg{}, errors.New(msg)
	case !ok:
		msg := fmt.Sprintf("Error: Unknown L2T type %d", msgType)
		return Msg{}, errors.New(msg)
	case observedLen != claimedLen:
		msg := fmt.Sprintf("Error: Incorrect L2T message length (header: %d, observed: %d", observedLen, claimedLen)
		return Msg{}, errors.New(msg)
	}
}
