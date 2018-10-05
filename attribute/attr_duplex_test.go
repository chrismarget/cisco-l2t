package attribute

import (
	"reflect"
	"strings"
	"testing"
)

func TestStringDuplex(t *testing.T) {
	attrTypesToTest := getAttrsByCategory(duplexCategory)
	for _, v := range attrTypesToTest {
		data1 := Attr{
			AttrType: v,
			AttrData: []byte{0},
		}
		expected1 := portDuplexToString[autoDuplex]
		result1, err := data1.String()
		if err != nil {
			t.Error(err)
		}
		if result1 != expected1 {
			t.Errorf("expected '%s', got '%s'", expected1, result1)
		}

		data2 := Attr{
			AttrType: v,
			AttrData: []byte{1},
		}
		expected2 := portDuplexToString[halfDuplex]
		result2, err := data2.String()
		if err != nil {
			t.Error(err)
		}
		if result2 != expected2 {
			t.Errorf("expected '%s', got '%s'", expected2, result2)
		}

		data3 := Attr{
			AttrType: v,
			AttrData: []byte{2},
		}
		expected3 := portDuplexToString[fullDuplex]
		result3, err := data3.String()
		if err != nil {
			t.Error(err)
		}
		if result3 != expected3 {
			t.Errorf("expected '%s', got '%s'", expected3, result3)
		}

		data4 := Attr{
			AttrType: v,
			AttrData: []byte{3},
		}
		_, err = data4.String()
		if err == nil {
			t.Error("Bogus duplex value should have produced an error")
		}

		data5 := Attr{
			AttrType: v,
			AttrData: []byte{0, 0},
		}
		_, err = data5.String()
		if err == nil {
			t.Error("Overlength duplex value should have produced an error")
		}

		data6 := Attr{
			AttrType: v,
			AttrData: []byte{},
		}
		_, err = data6.String()
		if err == nil {
			t.Error("Empty duplex value should have produced an error")
		}
	}
}

func TestNewDuplexAttr(t *testing.T) {
	attrTypesToTest := getAttrsByCategory(duplexCategory)
	for _, testType := range attrTypesToTest {
		var testPayload attrPayload
		var expected Attr
		var err error
		var result Attr

		for testDuplexVal, testString := range portDuplexToString {

			testPayload = attrPayload{stringData: strings.ToUpper(testString)}
			expected = Attr{
				AttrType: testType,
				AttrData: []byte{byte(testDuplexVal)},
			}
			result, err = NewAttr(testType, testPayload)
			if err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(result, expected) {
				t.Error("Unexpected result in TestNewDuplexAttr() upper case test. Structures don't match")

			}

			testPayload = attrPayload{stringData: strings.ToLower(testString)}
			expected = Attr{
				AttrType: testType,
				AttrData: []byte{byte(testDuplexVal)},
			}
			result, err = NewAttr(testType, testPayload)
			if err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(result, expected) {
				t.Error("Unexpected result in TestNewDuplexAttr() lower case test. Structures don't match")
			}
		}

		testInts := []int{0, 1, 2}
		for _, i := range testInts {
			testPayload = attrPayload{intData: i}
			expected = Attr{
				AttrType: testType,
				AttrData: []byte{byte(i)},
			}
			result, err = NewAttr(testType, testPayload)
			if err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(result, expected) {
				t.Error("Unexpected result in TestNewDuplexAttr() integer test. Structures don't match")
			}
		}
	}
}
