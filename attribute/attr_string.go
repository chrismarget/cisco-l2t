package attribute

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
)

const (
	stringTerminator = byte('\x00') // strings are null-teriminated
	maxStringWithoutTerminator = 252
)

type stringAttribute struct {
	attrType attrType
	attrData []byte
}

func (o stringAttribute) Type() attrType {
	return o.attrType
}

func (o stringAttribute) Len() uint8 {
	return uint8(TLsize + len(o.attrData))
}

func (o stringAttribute) String() string {
	return string(o.attrData[:len(o.attrData)-1])
}

func (o stringAttribute) Validate() error {
	err := checkTypeLen(o, stringCategory)
	if err != nil {
		return err
	}

	// Underlength?
	if int(o.Len()) < TLsize+len(string(stringTerminator)) {
		return fmt.Errorf("underlength string: got %d bytes (min %d)", o.Len(), TLsize+len(string(stringTerminator)))
	}

	// Overlength?
	if o.Len() > math.MaxUint8 {
		return fmt.Errorf("overlength string: got %d bytes (max %d)", o.Len(), math.MaxUint8)
	}

	// Ends with string terminator?
	if !strings.HasSuffix(string(o.attrData), string(stringTerminator)) {
		return fmt.Errorf("string missing termination character ` %#x'", string(stringTerminator))

	}

	// Printable?
	for _, v := range o.attrData[:(len(o.attrData) - 1)] {
		if v > unicode.MaxASCII || !unicode.IsPrint(rune(v)) {
			return errors.New("string is not printable.")
		}
	}

	return nil
}

func (o stringAttribute) Bytes() []byte {
	return o.attrData
}

// newsStringAttribute returns a new attribute from stringCategory
func (o *defaultAttrBuilder) newStringAttribute() (Attribute, error) {
	var stringBytes []byte
	switch {
	case o.bytesHasBeenSet:
		stringBytes = o.bytesPayload
	case o.intHasBeenSet:
		stringBytes = []byte(strconv.Itoa(int(o.intPayload)) + string(stringTerminator))
	case o.stringHasBeenSet:
		stringBytes = []byte(o.stringPayload + string(stringTerminator))
	default:
		return nil, fmt.Errorf("cannot build, no attribute payload found for category %s attribute", attrCategoryString[stringCategory])
	}

	a := stringAttribute {
		attrType: o.attrType,
		attrData: stringBytes,
	}

	err := a.Validate()
	if err != nil {
		return nil, err
	}

	return a, nil
}
