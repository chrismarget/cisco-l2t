package attribute

import (
	"fmt"
	"github.com/getlantern/errors"
	"math"
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

const (
	replyStatusSuccess = "Success"
	//	replyStatusDstNotFound = "Destination Mac address not found"
	//	replyStatusSrcNotFound = "Source Mac address not found"
	replyStatusUnknown = "Status Unknown"

	// The following strings were found together by strings-ing an IOS image.
	// Leap of faith makes me think they're reply status attribute messages.
	replyStatusInvalidMac           = "Invalid mac address"
	replyStatusInvalidIP            = "Invalid ip address"
	replyStausMultipleVlan          = "Mac found on multiple vlans"
	replyStatusDifferentVlan        = "Source and destination macs are on different vlans"
	replyStatusMultipleNeighbors    = "Device has multiple CDP neighbours"
	replyStatusNoNeighborIP         = "CDP neighbour has no ip"
	replyStatusNoNeighbor           = "No CDP neighbour"
	replyStatusSrcNotFound          = "Source Mac address not found"      // l2t attr type 0x0f data 0x07
	replyStatusDstNotFound          = "Destination Mac address not found" // l2t attr type 0x0f data 0x08
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
		1: replyStatusSuccess,
		7: replyStatusSrcNotFound,
		8: replyStatusDstNotFound,
	}
)

// stringifyReplyStatus takes an Attr belonging to replyStatusCategory, string-ifys it.
func stringifyReplyStatus(a Attr) (string, error) {
	var err error
	err = checkAttrInCategory(a, replyStatusCategory)
	if err != nil {
		return "", err
	}

	err = a.checkLen()
	if err != nil {
		return "", err
	}

	if msg, ok := replyStatusToString[replyStatus(a.AttrData[0])]; ok {
		return fmt.Sprintf("%s (%d)", msg, int(a.AttrData[0])), nil

	}

	return fmt.Sprintf("%s (%d)", replyStatusUnknown, int(a.AttrData[0])), nil
}

// newReplyStatusAttr returns an Attr with AttrType t and AttrData populated based on
// input payload. Input options are:
// - stringData (first choice, parses the string)
// - intData (second choice, value used directly)
func newReplyStatusAttr(t attrType, p attrPayload) (Attr, error) {
	result := Attr{AttrType: t}

	switch {
	case p.stringData != "":
		for k, v := range replyStatusToString {
			if strings.ToLower(p.stringData) == strings.ToLower(v) {
				result.AttrData = []byte{byte(k)}
				return result, nil
			}
		}
	case p.intData >= 0 && p.intData < math.MaxUint8:
		result.AttrData = []byte{byte(p.intData)}
		return result, nil
	}
	return Attr{}, errors.New("Error creating reply status attribute, no appropriate data supplied.")
}

// validateReplyStatus checks the AttrType and AttrData against norms for
// ReplyStatus type attributes.
func validateReplyStatus(a Attr) error {
	if attrCategoryByType[a.AttrType] != replyStatusCategory {
		msg := fmt.Sprintf("Attribute type %d cannot be validated against reply status criteria.", a.AttrType)
		return errors.New(msg)
	}
	return nil
}
