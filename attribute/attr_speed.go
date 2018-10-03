package attribute

import (
	"encoding/binary"
	"math"
	"strconv"
)

const (
	autoSpeedString = "Auto"
)

func stringSpeed(a attr) (string, error) {
	var err error
	err = checkAttrInCategory(a, speedCategory)
	if err != nil {
		return "", err
	}

	err = a.checkLen()
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
