package attribute

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"runtime"
	"strconv"
)

const (
	autoSpeedString = "Auto"
)

func stringSpeed(a attr) (string, error) {
	pc, _, _, _ := runtime.Caller(0)
	fname := runtime.FuncForPC(pc).Name()

	if attrCategoryByType[a.attrType] != speedCategory {
		msg := fmt.Sprintf("Cannot use %s on attribute with type %d.", fname, a.attrType)
		return "", errors.New(msg)
	}

	err := a.checkLen()
	if err != nil {
		return "", err
	}

	speedVal := binary.BigEndian.Uint32(a.attrData)
	if speedVal == 0 {
		return autoSpeedString, nil
	}

	speedUnits := "Mb/s"
	if speedVal >= 3 {
		speedUnits = "Gb/s"
		speedVal -= 3
	}

	return strconv.Itoa(int(math.Pow(10, float64(speedVal)))) + speedUnits, nil
}
