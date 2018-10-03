package attribute

import (
	"strconv"
)

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

	switch a.attrData[0] {
	case 1:
		return strconv.Itoa(int(a.attrData[0])) + " Success", nil
	case 9:
		return strconv.Itoa(int(a.attrData[0])) + " No CDP Neighbor", nil
	default:
		return strconv.Itoa(int(a.attrData[0])) + " Unknown", nil
	}
}
