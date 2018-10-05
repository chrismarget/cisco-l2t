package attribute

import (
	"encoding/binary"
	"github.com/getlantern/errors"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"
)

const (
	autoSpeedString = "Auto"
)

// stringSpeed takes an attribute, returns a nicely formatted string.
func stringSpeed(a Attr) (string, error) {
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

	// Default speed units is "Mb/s". speedVal is logarithmic, so large
	// values should switch units and decrement value accordingly
	var speedUnits string
	switch {
	case speedVal >= 3 && speedVal < 6:
		speedUnits = "Gb/s"
		speedVal -= 3
	case speedVal >= 6:
		speedUnits = "Tb/s"
		speedVal -= 0
	default:
		speedUnits = "Mb/s"
	}

	return strconv.Itoa(int(math.Pow(10, float64(speedVal)))) + speedUnits, nil
}

// newSpeedAttr returns an Attr with attrType t and attrData populated based on
// input payload. Input options are:
//
//   stringData (first choice)
//     If present, we parse the string
//
//   intData (second choice)
//     Value of 0 returns zeros/auto speed.
//     Values 0 < val < 10 are used in the way the l2t packet uses them: Speed is 10^value.
//     Values >= 10 are assumed to be Mb/s)
func newSpeedAttr(t attrType, p attrPayload) (Attr, error) {
	result := Attr{attrType: t}

	switch {
	// The zero case isn't handled here because it's the default for type int.
	// We consider the zero case last
	case p.intData > 0 && p.intData < 10:
		result.attrData = make([]byte, 4)
		binary.BigEndian.PutUint32(result.attrData, uint32(p.intData))
		return result, nil
	case p.intData >= 10:
		speedOut := math.Log10(float64(p.intData))
		if speedOut > math.MaxUint32 {
			return Attr{}, errors.New("Error: Speed value out of range.")
		}
		result.attrData = make([]byte, 4)
		binary.BigEndian.PutUint32(result.attrData, uint32(speedOut))
		return result, nil
	case p.stringData != "":
		inString := strings.ToLower(p.stringData)
		var onlyDigits = regexp.MustCompile(`^[0-9]+$`).MatchString
		if onlyDigits(inString) {
			in, err := strconv.Atoi(inString)
			if err != nil {
				return Attr{}, err
			}
			return newSpeedAttr(t, attrPayload{intData: in})
		}

		var suffixToMultiplier = map[string]int{
			"auto": 0,
			"mb/s": 1,
			"mbps": 1,
			"mbs":  1,
			"mb":   1,
			"gb/s": 1000,
			"gbps": 1000,
			"gbs":  1000,
			"gb":   1000,
			"tb/s": 1000000,
			"tbps": 1000000,
			"tbs":  1000000,
			"tb":   1000000,
		}

		// Loop over suffixes, see if the supplied string matches one.
		for suffix, multiplier := range suffixToMultiplier {
			if strings.HasSuffix(inString, suffix) {
				// Found one! Trim the suffix and whitespace from the string.
				trimmed := strings.TrimSpace(strings.TrimSuffix(inString, suffix))

				// If all we got was "mb/s" or "auto", set the speed value to
				// "0" so we can do math on it later.
				if trimmed == "" {
					trimmed = "0"
				}

				// Now make sure we've only got math-y characters left in the string.
				if strings.Trim(trimmed, "0123456789.") != "" {
					return Attr{}, errors.New("Error creating speed attribute, unable to parse string.")
				}

				// At the end of this section, we'll have turned "10gb/s" into 10000000,
				// stored it in speedVal.
				var speedVal float64
				var err error
				if speedVal, err = strconv.ParseFloat(trimmed, 32); err != nil {
					log.Println(err.Error())
					return Attr{}, err
				}

				// Now that it's just a numeric value, recurse this function
				// with "intData" payload type.
				return newSpeedAttr(t, attrPayload{intData: int(speedVal) * multiplier})
			}
		}
	case p.intData == 0:
		result.attrData = make([]byte, 4)
		binary.BigEndian.PutUint32(result.attrData, uint32(p.intData))
		return result, nil
	}
	return Attr{}, errors.New("Error creating speed attribute, no appropriate data supplied.")
}
