package attribute

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

const (
	autoSpeedString        = "Auto"
	megaBitPerSecondSuffix = "Mbps"
	maxSpeedMbps           = 100000000
	maxSpeedWireFormat     = 9
)

type speedAttribute struct {
	attrType AttrType
	attrData []byte
}

func (o speedAttribute) Type() AttrType {
	return o.attrType
}

func (o speedAttribute) Len() uint8 {
	return uint8(TLsize + len(o.attrData))
}

func (o speedAttribute) String() string {
	// 32-bit zero is a special case
	if reflect.DeepEqual(o.attrData, []byte{0, 0, 0, 0}) {
		return autoSpeedString
	}

	return SpeedBytesToString(o.attrData)
}

func (o speedAttribute) Validate() error {
	err := checkTypeLen(o, speedCategory)
	if err != nil {
		return err
	}

	if binary.BigEndian.Uint32(o.attrData) > maxSpeedWireFormat {
		return fmt.Errorf("wire format speed `%d' exceeds maximum value (%d)", binary.BigEndian.Uint32(o.attrData), maxSpeedWireFormat)
	}

	speedVal := int(math.Pow(10, float64(binary.BigEndian.Uint32(o.attrData))))
	if speedVal > maxSpeedMbps {
		return fmt.Errorf("interface speed `%d' exceeds maximum value (%d)", speedVal, maxSpeedMbps)
	}

	return nil
}

func (o speedAttribute) Bytes() []byte {
	return o.attrData
}

// newSpeedAttribute returns a new attribute from speedCategory
func (o *defaultAttrBuilder) newSpeedAttribute() (Attribute, error) {
	var err error
	var speedBytes []byte
	switch {
	case o.stringHasBeenSet:
		speedBytes, err = SpeedStringToBytes(o.stringPayload)
		if err != nil {
			return nil, err
		}
	case o.bytesHasBeenSet:
		speedBytes = o.bytesPayload
	case o.intHasBeenSet:
		speedBytes, err = SpeedStringToBytes(strconv.Itoa(int(o.intPayload)) + megaBitPerSecondSuffix)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("cannot build, no attribute payload found for category %s attribute", attrCategoryString[speedCategory])
	}

	a := &speedAttribute{
		attrType: o.attrType,
		attrData: speedBytes,
	}

	err = a.Validate()
	if err != nil {
		return nil, err
	}

	return a, nil
}

// SpeedBytesToString takes an input speed in wire format,
// returns a string.
//
// {0,0,0,0} -> "Auto"
//
// {0,0,0,1} -> 10Mb/s
//
// {0,0,0,4} -> 10Gb/s
func SpeedBytesToString(b []byte) string {
	// Default speed units is "Mb/s". Value of speedVal is logarithmic.
	// With large values we switch units and decrement the value accordingly.
	speedVal := binary.BigEndian.Uint32(b)
	var speedUnits string
	switch {
	case speedVal >= 3 && speedVal < 6:
		speedUnits = "Gb/s"
		speedVal -= 3
	case speedVal >= 6:
		speedUnits = "Tb/s"
		speedVal -= 6
	default:
		speedUnits = "Mb/s"
	}

	return strconv.Itoa(int(math.Pow(10, float64(speedVal)))) + speedUnits
}

// SpeedStringToBytes takes an input string, returns
// speed as an Uint32 expressed in Mb/s
//
// Auto -> 0
//
// 10Mb/s -> 10
//
// 10Gb/s -> 10000
func SpeedStringToBytes(s string) ([]byte, error) {
	// Did we get valid characters?
	for _, v := range s {
		if v > unicode.MaxASCII || !unicode.IsPrint(rune(v)) {
			return nil, errors.New("string contains invalid characters")
		}
	}

	// normalize instring
	s = strings.TrimSpace(strings.ToLower(s))

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
		if strings.HasSuffix(s, suffix) {
			// Found one! Trim the suffix and whitespace from the string.
			trimmed := strings.TrimSpace(strings.TrimSuffix(s, suffix))

			// If all we got was a suffix (maybe "mb/s", more likely "auto"),
			// set the speed value to "0" so we can do math on it later.
			if trimmed == "" {
				trimmed = "0"
			}

			// Now make sure we've only got math-y characters left in the string.
			if strings.Trim(trimmed, "0123456789.") != "" {
				return nil, fmt.Errorf("cannot parse `%s' as an interface speed", s)
			}

			// speedFloat will contain the value previously expressed by
			// the passed string, ignoring the suffix:
			//
			// If "100Gb/s", speedFloat will be float64(100)
			speedFloat, err := strconv.ParseFloat(trimmed, 32)
			if err != nil {
				return nil, err
			}

			// speedFloatMbps will contain the value previously expressed by
			// the passed string, corrected to account for the suffix:
			//
			// If "100Gb/s", speedFloat will be float64(100000)
			speedFloatMbps := speedFloat * float64(multiplier)

			// speedBytes will contain the value from the passed string
			// expressed in wire format.
			speedBytes := make([]byte, 4)
			binary.BigEndian.PutUint32(speedBytes, uint32(math.Log10(speedFloatMbps)))

			return speedBytes, nil
		}
	}

	return nil, fmt.Errorf("cannot parse speed string: `%s'", s)
}
