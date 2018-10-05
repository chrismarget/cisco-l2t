package attribute

import (
	"log"
	"reflect"
	"testing"
)

func TestStringString(t *testing.T) {
	attrTypesToTest := getAttrsByCategory(stringCategory)
	for _, v := range attrTypesToTest {
		data1 := attr{
			attrType: v,
			attrData: []byte{65, 0},
		}
		expected1 := "A"
		result1, err := data1.String()
		if err != nil {
			t.Error(err)
		}
		if result1 != expected1 {
			t.Errorf("expected '%s', got '%s'", expected1, result1)
		}
	}
}

func TestNewStringAttr(t *testing.T) {
	attrTypesToTest := getAttrsByCategory(stringCategory)
	for _, testType := range attrTypesToTest {
		var stringsToTest []string
		stringsToTest = append(stringsToTest, "hello")

		var expectedResults []attr
		expectedResults = append(expectedResults, attr{attrType: testType, attrData: []byte{104, 101, 108, 108, 111, 0}})

		for k, _ := range stringsToTest {
			result, err := NewAttr(testType, attrPayload{stringData: stringsToTest[k]})
			if err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(result, expectedResults[k]) {
				t.Error("Error, attribute structs don't match.")
			}
		}

		var err error
		// Test with empty string
		_, err = NewAttr(testType, attrPayload{})
		if err == nil {
			t.Error("Empty string should have produced an error.")
		}

		// Test with non-printables string
		_, err = NewAttr(testType, attrPayload{stringData: string(255)})
		if err == nil {
			t.Error("Empty string should have produced an error.")
		}

		var p attrPayload
		for i := 0; i < 253; i++ {
			p.stringData = p.stringData + "A"
		}
		log.Println(len(p.stringData))
		_, err = NewAttr(testType, p)
		if err == nil {
			t.Error("Over length string should have produced an error.")
		}
	}
}
