package attribute

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestStringStatus(t *testing.T) {
	for i := 0; i <= 255; i++ {
		data := attr{
			attrType: replyStatusType,
			attrData: []byte{byte(i)},
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
		stringsToTest = append(stringsToTest, replyStatusNoCDPNeighbor)
		stringsToTest = append(stringsToTest, strings.ToUpper(replyStatusSuccess))
		stringsToTest = append(stringsToTest, strings.ToUpper(replyStatusNoCDPNeighbor))
		stringsToTest = append(stringsToTest, strings.ToLower(replyStatusSuccess))
		stringsToTest = append(stringsToTest, strings.ToLower(replyStatusNoCDPNeighbor))

		var testPayload []attrPayload
		for _, testString := range stringsToTest {
			testPayload = append(testPayload, attrPayload{stringData: testString})
		}

		var expectedResult []attr
		for _, v := range stringsToTest {
			for i, j := range replyStatusToString {
				if strings.ToLower(j) == strings.ToLower(v) {
					expectedResult = append(expectedResult, attr{testType, []byte{byte(i)}})
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
