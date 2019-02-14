package attribute

import (
	"bytes"
	"fmt"
	"math"
	"strings"
	"testing"
)

func TestDuplexAttribute_String(t *testing.T) {
	var (
		duplexStringTestData = map[string]portDuplex{
			"Auto": portDuplex(0),
			"Half": portDuplex(1),
			"Full": portDuplex(2),
		}
	)

	for _, duplexAttrType := range getAttrsByCategory(duplexCategory) {
		for expected, data := range duplexStringTestData {
			testAttr := duplexAttribute{
				attrType: duplexAttrType,
				attrData: []byte{byte(data)},
			}
			result := testAttr.String()
			if result != expected {
				t.Fatalf("expected %s, got %s", expected, result)
			}
		}
	}
}

func TestDuplexAttribute_Validate_WithGoodData(t *testing.T) {
	goodData := [][]byte{
		[]byte{0},
		[]byte{1},
		[]byte{2},
	}
	for _, duplexAttrType := range getAttrsByCategory(duplexCategory) {
		for _, testData := range goodData {
			testAttr := duplexAttribute{
				attrType: duplexAttrType,
				attrData: testData,
			}
			err := testAttr.Validate()
			if err != nil {
				t.Fatalf(err.Error()+"\n"+"Supposed good data %s produced error for %s.",
					fmt.Sprintf("%v", []byte(testAttr.attrData)), attrTypeString[duplexAttrType])
			}
		}
	}
}

func TestDuplexAttribute_Validate_WithBadData(t *testing.T) {
	badData := [][]byte{
		nil,
		[]byte{},
		[]byte{0, 0},
	}

	for i := 3; i <= math.MaxUint8; i++ {
		badData = append(badData, []byte{byte(i)})
	}

	for _, duplexAttrType := range getAttrsByCategory(duplexCategory) {
		for _, testData := range badData {
			testAttr := duplexAttribute{
				attrType: duplexAttrType,
				attrData: testData,
			}

			err := testAttr.Validate()
			if err == nil {
				t.Fatalf("Bad data %s in %s did not error.",
					fmt.Sprintf("%v", []byte(testAttr.attrData)), attrTypeString[duplexAttrType])
			}
		}
	}
}

func TestNewAttrBuilder_Duplex(t *testing.T) {
	for _, duplexAttrType := range getAttrsByCategory(duplexCategory) {
		for k, v := range portDuplexToString {
			expected := []byte{byte(duplexAttrType), 3, byte(k)}
			byInt, err := NewAttrBuilder().SetType(duplexAttrType).SetInt(uint32(k)).Build()
			if err != nil {
				t.Fatal(err)
			}
			byString, err := NewAttrBuilder().SetType(duplexAttrType).SetString(v).Build()
			if err != nil {
				t.Fatal(err)
			}
			byStringLower, err := NewAttrBuilder().SetType(duplexAttrType).SetString(strings.ToLower(v)).Build()
			if err != nil {
				t.Fatal(err)
			}
			byStringUpper, err := NewAttrBuilder().SetType(duplexAttrType).SetString(strings.ToUpper(v)).Build()
			if err != nil {
				t.Fatal(err)
			}
			byByte, err := NewAttrBuilder().SetType(duplexAttrType).SetBytes([]byte{byte(k)}).Build()
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Compare(expected, MarshalV1Attribute(byInt)) != 0 {
				t.Fatal("Attributes don't match")
			}
			if bytes.Compare(byInt.Bytes(), byString.Bytes()) != 0 {
				t.Fatal("Attributes don't match")
			}
			if bytes.Compare(byString.Bytes(), byStringLower.Bytes()) != 0 {
				t.Fatal("Attributes don't match")
			}
			if bytes.Compare(byStringLower.Bytes(), byStringUpper.Bytes()) != 0 {
				t.Fatal("Attributes don't match")
			}
			if bytes.Compare(byStringUpper.Bytes(), byByte.Bytes()) != 0 {
				t.Fatal("Attributes don't match")
			}
			if bytes.Compare(byByte.Bytes(), byInt.Bytes()) != 0 {
				t.Fatal("Attributes don't match")
			}
		}
	}
}
