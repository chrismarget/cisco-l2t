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
	for _, replyStatusAttrType := range  getAttrsByCategory(replyStatusCategory) {
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

//
//func TestStringStatus(t *testing.T) {
//	for i := 0; i <= 255; i++ {
//		data := Attr{
//			AttrType: replyStatusType,
//			AttrData: []byte{byte(i)},
//		}
//
//		var expected string
//		switch i {
//		case 1:
//			expected = "Success (1)"
//		case 7:
//			expected = "Source Mac address not found (7)"
//		case 8:
//			expected = "Destination Mac address not found (8)"
//		default:
//			expected = "Status Unknown (" + strconv.Itoa(i) + ")"
//		}
//
//		result, err := data.String()
//		if err != nil {
//			t.Error(err)
//		}
//		if result != expected {
//			t.Errorf("expected '%s', got '%s'", expected, result)
//		}
//	}
//}
//
//func TestNewReplyStatusAttrWithString(t *testing.T) {
//	attrTypesToTest := getAttrsByCategory(replyStatusCategory)
//	for _, testType := range attrTypesToTest {
//		var stringsToTest []string
//		stringsToTest = append(stringsToTest, replyStatusSuccess)
//		stringsToTest = append(stringsToTest, replyStatusSrcNotFound)
//		stringsToTest = append(stringsToTest, replyStatusDstNotFound)
//		stringsToTest = append(stringsToTest, strings.ToUpper(replyStatusSuccess))
//		stringsToTest = append(stringsToTest, strings.ToUpper(replyStatusSrcNotFound))
//		stringsToTest = append(stringsToTest, strings.ToUpper(replyStatusDstNotFound))
//		stringsToTest = append(stringsToTest, strings.ToLower(replyStatusSuccess))
//		stringsToTest = append(stringsToTest, strings.ToLower(replyStatusSrcNotFound))
//		stringsToTest = append(stringsToTest, strings.ToLower(replyStatusDstNotFound))
//
//		var testPayload []attrPayload
//		for _, testString := range stringsToTest {
//			testPayload = append(testPayload, attrPayload{stringData: testString})
//		}
//
//		var expectedResult []Attr
//		for _, v := range stringsToTest {
//			for i, j := range replyStatusToString {
//				if strings.ToLower(j) == strings.ToLower(v) {
//					expectedResult = append(expectedResult, Attr{testType, []byte{byte(i)}})
//				}
//			}
//		}
//
//		for i, p := range testPayload {
//			result, err := NewAttr(testType, p)
//			if err != nil {
//				t.Error(err)
//			}
//			if !reflect.DeepEqual(result, expectedResult[i]) {
//				t.Error("Error: Structures do not match")
//			}
//		}
//	}
//}
//
//func TestNewReplyStatusAttrWithBogusString(t *testing.T) {
//	attrTypesToTest := getAttrsByCategory(replyStatusCategory)
//	for _, testType := range attrTypesToTest {
//		_, err := NewAttr(testType, attrPayload{stringData: "bogus"})
//		if err == nil {
//			t.Error("Bogus string should have produced an error.")
//		}
//	}
//
//}
//
//func TestValidateReplyStatus(t *testing.T) {
//	replyStatusTypes := map[attrType]bool{}
//	for i := 0; i <= 255; i++ {
//		if attrCategoryByType[attrType(i)] == replyStatusCategory {
//			replyStatusTypes[attrType(i)] = true
//		}
//	}
//
//	validReplyStatuses := map[replyStatus]bool{}
//	for i := 0; i <= 255; i++ {
//		if _, ok := replyStatusToString[replyStatus(i)]; ok {
//			validReplyStatuses[replyStatus(i)] = true
//		}
//	}
//
//	// Loop over attribte types to test
//	for at := 0; at <= 255; at++ {
//		// Loop over reply statuses to test
//		for rs := 0; rs <= 255; rs++ {
//			a := Attr{AttrType: attrType(at), AttrData: []byte{byte(rs)}}
//			err := validateReplyStatus(a)
//			switch {
//			case replyStatusTypes[attrType(at)] && validReplyStatuses[replyStatus(rs)] && err != nil:
//				msg := fmt.Sprintf("Attribute type %d with status code %d should not produce ReplyStatus validation errors: %s", at, rs, err)
//				t.Error(msg)
//			case !replyStatusTypes[attrType(at)] && err == nil:
//				msg := fmt.Sprintf("Attribute type %d should have produced ReplyStatus validation errors.", at)
//				t.Error(msg)
//			case !validReplyStatuses[replyStatus(rs)] && err == nil:
//				msg := fmt.Sprintf("Reply Status code %d should have produced ReplyStatus validation errors.", rs)
//				t.Error(msg)
//			}
//		}
//	}
//}
