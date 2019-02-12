package attribute

import (
	"bytes"
	"fmt"
	"math"
	"strings"
	"testing"
)

func TestReplyStatusAttribute_String(t *testing.T) {
	replyStatusStringTestData := make(map[byte]string)

	// Preload all test data with "Status Unknown (<val>)"
	for i := 0; i <= math.MaxUint8; i++ {
		replyStatusStringTestData[byte(i)] = fmt.Sprintf("Status unknown (%d)", i)
	}

	// Some of the preloaded test data has actual values. Fix 'em.
	replyStatusStringTestData[1] = "Success"
	replyStatusStringTestData[7] = "Source Mac address not found"
	replyStatusStringTestData[8] = "Destination Mac address not found"

	for _, replyStatusAttrType := range getAttrsByCategory(replyStatusCategory) {
		for data, expected := range replyStatusStringTestData {
			testAttr := replyStatusAttribute{
				attrType: replyStatusAttrType,
				attrData: []byte{data},
			}
			result := testAttr.String()
			if result != expected {
				t.Fatalf("expected %s, got %s", expected, result)
			}
		}
	}
}

func TestReplyStatusAttribute_Validate_WithGoodData(t *testing.T) {
	var goodData [][]byte
	// test all possible values
	for i := 1; i <= math.MaxUint8; i++ {
		goodData = append(goodData, []byte{byte(i)})
	}

	for _, replyStatusAttrType := range getAttrsByCategory(replyStatusCategory) {
		for _, testData := range goodData {
			testAttr := replyStatusAttribute{
				attrType: replyStatusAttrType,
				attrData: testData,
			}
			err := testAttr.Validate()
			if err != nil {
				t.Fatalf(err.Error()+"\n"+"Supposed good data %s produced error for %s.",
					fmt.Sprintf("%v", []byte(testAttr.attrData)), attrTypeString[replyStatusAttrType])
			}
		}
	}
}

func TestReplyStatusAttribute_Validate_WithBadData(t *testing.T) {
	goodData := [][]byte{
		nil,
		[]byte{},
		[]byte{0, 0},
	}

	for _, replyStatusAttrType := range getAttrsByCategory(replyStatusCategory) {
		for _, testData := range goodData {
			testAttr := replyStatusAttribute{
				attrType: replyStatusAttrType,
				attrData: testData,
			}

			err := testAttr.Validate()
			if err == nil {
				t.Fatalf("Bad data %s in %s did not error.",
					fmt.Sprintf("%v", []byte(testAttr.attrData)), attrTypeString[replyStatusAttrType])
			}
		}
	}
}

func TestNewAttrBuilder_ReplyStatus(t *testing.T) {
	for _, replyStatusAttrType := range getAttrsByCategory(replyStatusCategory) {
		for k, v := range replyStatusToString {
			expected := []byte{byte(replyStatusAttrType), 3, byte(k)}
			byInt, err := NewAttrBuilder().SetType(replyStatusAttrType).SetInt(uint32(k)).Build()
			if err != nil {
				t.Fatal(err)
			}
			byString, err := NewAttrBuilder().SetType(replyStatusAttrType).SetString(v).Build()
			if err != nil {
				t.Fatal(err)
			}
			byStringLower, err := NewAttrBuilder().SetType(replyStatusAttrType).SetString(strings.ToLower(v)).Build()
			if err != nil {
				t.Fatal(err)
			}
			byStringUpper, err := NewAttrBuilder().SetType(replyStatusAttrType).SetString(strings.ToUpper(v)).Build()
			if err != nil {
				t.Fatal(err)
			}
			byByte, err := NewAttrBuilder().SetType(replyStatusAttrType).SetBytes([]byte{byte(k)}).Build()
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Compare(expected, MarshalAttribute(byInt)) != 0 {
				t.Fatal("Attributes don't match")
			}
			if bytes.Compare(byInt.Bytes(), byByte.Bytes()) != 0 {
				t.Fatal("Attributes don't match")
			}
			if bytes.Compare(byByte.Bytes(), byString.Bytes()) != 0 {
				t.Fatal("Attributes don't match")
			}
			if bytes.Compare(byString.Bytes(), byStringLower.Bytes()) != 0 {
				t.Fatal("Attributes don't match")
			}
			if bytes.Compare(byStringLower.Bytes(), byStringUpper.Bytes()) != 0 {
				t.Fatal("Attributes don't match")
			}
			if bytes.Compare(byStringUpper.Bytes(), byInt.Bytes()) != 0 {
				t.Fatal("Attributes don't match")
			}
		}
	}
}
