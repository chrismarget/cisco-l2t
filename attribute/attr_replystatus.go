package attribute

import (
	"errors"
	"fmt"
	"runtime"
	"strconv"
)

func stringReplyStatus(a attr) (string, error) {
	pc, _, _, _ := runtime.Caller(0)
	fname := runtime.FuncForPC(pc).Name()

	if attrCategoryByType[a.attrType] != statusCategory {
		msg := fmt.Sprintf("Cannot use %s on attribute with type %d.", fname, a.attrType)
		return "", errors.New(msg)
	}

	err := a.checkLen()
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
