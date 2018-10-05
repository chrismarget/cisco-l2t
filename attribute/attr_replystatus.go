package attribute

import (
	"fmt"
	"github.com/getlantern/errors"
	"math"
	"strings"
)

const (
	replyStatusSuccess       = "Success"
	replyStatusNoCDPNeighbor = "No CDP Neighbor"
	replyStatusUnknown       = "Status Unknown"
)

type (
	replyStatus byte
)

var (
	replyStatusToString = map[replyStatus]string{
		1: replyStatusSuccess,
		9: replyStatusNoCDPNeighbor,
	}
)

// stringReplyStatus takes an attr belonging to replyStatusCategory, string-ifys it.
func stringReplyStatus(a attr) (string, error) {
	var err error
	err = checkAttrInCategory(a, replyStatusCategory)
	if err != nil {
		return "", err
	}

	err = a.checkLen()
	if err != nil {
		return "", err
	}

	if msg, ok := replyStatusToString[replyStatus(a.attrData[0])]; ok {
		return fmt.Sprintf("%s (%d)", msg, int(a.attrData[0])), nil

	}

	return fmt.Sprintf("%s (%d)", replyStatusUnknown, int(a.attrData[0])), nil
}

// newReplyStatusAttr returns an attr with attrType t and attrData populated based on
// input payload. Input options are:
// - stringData (first choice, parses the string)
// - intData (second choice, value used directly)
func newReplyStatusAttr(t attrType, p attrPayload) (attr, error) {
	result := attr{attrType: t}

	switch {
	case p.stringData != "":
		for k, v := range replyStatusToString {
			if strings.ToLower(p.stringData) == strings.ToLower(v) {
				result.attrData = []byte{byte(k)}
				return result, nil
			}
		}
	case p.intData >= 0 && p.intData < math.MaxUint8:
		result.attrData = []byte{byte(p.intData)}
		return result, nil
	}
	return attr{}, errors.New("Error creating reply status, no appropriate data supplied.")
}
