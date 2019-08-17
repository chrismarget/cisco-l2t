package attribute

import (
	"fmt"
	"strings"
)

//Some stuff I found by strings-ing a binary:
//
//Multiple devices connected to %s port of %s [%i].
//Layer2 trace aborted.
//Invalid src/dest received from %s [%i].
//Layer2 trace aborted.
//Unable to locate port for src %e on %s [%i].
//Layer2 trace aborted.
//Unable to locate port for dst %e on %s [%i].
//Layer2 trace aborted.
//%e belongs to mutliple vlans on %s [%i].
//Layer2 trace aborted.
//Source and destination vlan mismatch discovered on %s [%i].
//Layer2 trace aborted.
//Failed to get any layer2 path information from %s [%i].
//Layer2 trace aborted.
//Layer2 path not through %s [%i].
//Layer2 trace aborted.
//Unknown return code %d from %s [%i].
//
//Invalid mac address
//Invalid ip address
//Mac found on multiple vlans
//Source and destination macs are on different vlans
//Device has multiple CDP neighbours
//CDP neighbour has no ip
//No CDP neighbour
//Source Mac address not found
//Destination Mac address not found
//Incorrect source interface specified
//Incorrect destination interface specified
//Device has Multiple CDP neighbours on source port
//Device has Multiple CDP neighbours on destination port

// observed reply status:
// 1 - normal trace, end of the line (no cdp neighbor?)
// 2 - normal trace, cdp neighbor returned
// 3 - bogus vlan (debugs say "internal error")
// 5 - multiple CDP neighbors
// 7 - source mac not found (with

const (
	ReplyStatusSuccess = "Success"
	//	ReplyStatusDstNotFound = "Destination Mac address not found"
	//	ReplyStatusSrcNotFound = "Source Mac address not found"
	ReplyStatusUnknown = "Status unknown"

	// The following strings were found together by strings-ing an IOS image.
	// Leap of faith makes me think they're reply status attribute messages.
	replyStatusInvalidMac           = "Invalid mac address"
	replyStatusInvalidIP            = "Invalid ip address"
	replyStausMultipleVlan          = "Mac found on multiple vlans"
	replyStatusDifferentVlan        = "Source and destination macs are on different vlans"
	ReplyStatusMultipleNeighbors    = "Device has multiple CDP neighbours"
	replyStatusNoNeighborIP         = "CDP neighbour has no ip"
	ReplyStatusNoNeighbor           = "No CDP neighbour"
	ReplyStatusSrcNotFound          = "Source Mac address not found"      // l2t attr type 0x0f data 0x07
	ReplyStatusDstNotFound          = "Destination Mac address not found" // l2t attr type 0x0f data 0x08
	replyStatusWrongSrcInterface    = "Incorrect source interface specified"
	replyStatusWrongDstInterface    = "Incorrect destination interface specified"
	replyStatusSrcMultipleNeighbors = "Device has Multiple CDP neighbours on source port"
	replyStatusDstMultipleNeighbors = "Device has Multiple CDP neighbours on destination port"
)

type (
	replyStatus byte
)

var (
	replyStatusToString = map[replyStatus]string{
		1: ReplyStatusSuccess,
		5: ReplyStatusMultipleNeighbors,
		7: ReplyStatusSrcNotFound,
		8: ReplyStatusDstNotFound,
		9: ReplyStatusNoNeighbor,
	}
)

type replyStatusAttribute struct {
	attrType AttrType
	attrData []byte
}

func (o replyStatusAttribute) Type() AttrType {
	return o.attrType
}

func (o replyStatusAttribute) Len() uint8 {
	return uint8(TLsize + len(o.attrData))
}

func (o replyStatusAttribute) String() string {
	if status, ok := replyStatusToString[replyStatus(o.attrData[0])]; ok {
		return status
	}
	return fmt.Sprintf("%s (%d)", ReplyStatusUnknown, o.attrData[0])
}

func (o replyStatusAttribute) Validate() error {
	err := checkTypeLen(o, replyStatusCategory)
	if err != nil {
		return err
	}
	return nil
}

func (o replyStatusAttribute) Bytes() []byte {
	return o.attrData
}

// newReplyStatusAttribute returns a new attribute from replyStatusCategory
func (o *defaultAttrBuilder) newReplyStatusAttribute() (Attribute, error) {
	var replyStatusByte byte
	var success bool
	switch {
	case o.stringHasBeenSet:
		for replyStatus, replyStatusString := range replyStatusToString {
			if strings.ToLower(o.stringPayload) == strings.ToLower(replyStatusString) {
				replyStatusByte = byte(replyStatus)
				success = true
			}
		}
		if !success {
			return nil, fmt.Errorf("string payload `%s' unrecognized for reply status type", o.stringPayload)
		}
	case o.intHasBeenSet:
		replyStatusByte = uint8(o.intPayload)
	case o.bytesHasBeenSet:
		if len(o.bytesPayload) != 1 {
			return nil, fmt.Errorf("cannot use %d bytes to build a reply status attribute", len(o.bytesPayload))
		}
		replyStatusByte = o.bytesPayload[0]
	default:
		return nil, fmt.Errorf("cannot build, no attribute payload found for category %s attribute", attrCategoryString[replyStatusCategory])
	}

	a := &replyStatusAttribute{
		attrType: o.attrType,
		attrData: []byte{replyStatusByte},
	}

	err := a.Validate()
	if err != nil {
		return nil, err
	}

	return a, nil
}
