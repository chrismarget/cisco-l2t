package attribute

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestStringStatus(t *testing.T) {
	for i := 0; i <= 255; i++ {
		data := Attr{
			AttrType: replyStatusType,
			AttrData: []byte{byte(i)},
		}

		var expected string
		switch i {
		case 1:
			expected = "Success (1)"
		case 9:
			expected = "No CDP Neighbor (9)"
		default:
			expected = "Status Unknown (" + strconv.Itoa(i) + ")"
		}

		result, err := data.String()
		if err != nil {
			t.Error(err)
		}
		if result != expected {
			t.Errorf("expected '%s', got '%s'", expected, result)
		}
	}
}

func TestNewReplyStatusAttrWithString(t *testing.T) {
	attrTypesToTest := getAttrsByCategory(replyStatusCategory)
	for _, testType := range attrTypesToTest {
		var stringsToTest []string
		stringsToTest = append(stringsToTest, replyStatusSuccess)
		stringsToTest = append(stringsToTest, replyStatusSrcNotFound)
		stringsToTest = append(stringsToTest, replyStatusDstNotFound)
		stringsToTest = append(stringsToTest, strings.ToUpper(replyStatusSuccess))
		stringsToTest = append(stringsToTest, strings.ToUpper(replyStatusSrcNotFound))
		stringsToTest = append(stringsToTest, strings.ToUpper(replyStatusDstNotFound))
		stringsToTest = append(stringsToTest, strings.ToLower(replyStatusSuccess))
		stringsToTest = append(stringsToTest, strings.ToLower(replyStatusSrcNotFound))
		stringsToTest = append(stringsToTest, strings.ToLower(replyStatusDstNotFound))

		var testPayload []attrPayload
		for _, testString := range stringsToTest {
			testPayload = append(testPayload, attrPayload{stringData: testString})
		}

		var expectedResult []Attr
		for _, v := range stringsToTest {
			for i, j := range replyStatusToString {
				if strings.ToLower(j) == strings.ToLower(v) {
					expectedResult = append(expectedResult, Attr{testType, []byte{byte(i)}})
				}
			}
		}

		for i, p := range testPayload {
			result, err := NewAttr(testType, p)
			if err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(result, expectedResult[i]) {
				t.Error("Error: Structures do not match")
			}
		}
	}
}

func TestNewReplyStatusAttrWithBogusString(t *testing.T) {
	attrTypesToTest := getAttrsByCategory(replyStatusCategory)
	for _, testType := range attrTypesToTest {
		_, err := NewAttr(testType, attrPayload{stringData: "bogus"})
		if err == nil {
			t.Error("Bogus string should have produced an error.")
		}
	}

}

func TestValidateReplyStatus(t *testing.T) {
	replyStatusTypes := map[attrType]bool{}
	for i := 0; i <= 255; i++ {
		if attrCategoryByType[attrType(i)] == replyStatusCategory {
			replyStatusTypes[attrType(i)] = true
		}
	}

	validReplyStatuses := map[replyStatus]bool{}
	for i := 0; i <= 255; i++ {
		if _, ok := replyStatusToString[replyStatus(i)]; ok {
			validReplyStatuses[replyStatus(i)] = true
		}
	}

	// Loop over attribte types to test
	for at := 0; at <= 255; at++ {
		// Loop over reply statuses to test
		for rs := 0; rs <= 255; rs++ {
			a := Attr{AttrType: attrType(at), AttrData: []byte{byte(rs)}}
			err := validateReplyStatus(a)
			switch {
			case replyStatusTypes[attrType(at)] && validReplyStatuses[replyStatus(rs)] && err != nil:
				msg := fmt.Sprintf("Attribute type %d with status code %d should not produce ReplyStatus validation errors: %s", at, rs, err)
				t.Error(msg)
			case !replyStatusTypes[attrType(at)] && err == nil:
				msg := fmt.Sprintf("Attribute type %d should have produced ReplyStatus validation errors.", at)
				t.Error(msg)
			case !validReplyStatuses[replyStatus(rs)] && err == nil:
				msg := fmt.Sprintf("Reply Status code %d should have produced ReplyStatus validation errors.", rs)
				t.Error(msg)
			}
		}
	}
}
