package attribute

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestStringIPv4(t *testing.T) {
	attrTypesToTest := getAttrsByCategory(ipv4Category)
	for _, v := range attrTypesToTest {
		data1 := attr{
			attrType: v,
			attrData: []byte{192, 168, 10, 11},
		}
		expected1 := "192.168.10.11"
		result1, err := data1.String()
		if err != nil {
			t.Error(err)
		}
		if result1 != expected1 {
			t.Errorf("expected '%s', got '%s'", expected1, result1)
		}
	}
}

// todo: test ipaddr and int types
func TestNewIPv4Attr(t *testing.T) {
	attrTypesToTest := getAttrsByCategory(ipv4Category)
	for _, testType := range attrTypesToTest {
		stringsToTest := []string{"0.0.0.0", "192.168.1.2", "255.255.255.255"}
		for _, testString := range stringsToTest {
			var testPayload attrPayload
			var expected attr
			var err error
			var result attr

			// building the "expected" structure requires us to walk the test
			// string the hard way.
			octets := strings.Split(testString, ".")
			var data []byte
			for _, o := range octets{
				var i int
				i, err = strconv.Atoi(o)
				data = append(data, byte(i))
			}
			expected = attr{
				attrType: testType,
				attrData: data,
			}

			testPayload.stringData = testString
			result, err = NewAttr(testType, testPayload)
			if err != nil {
				t.Error(err)
			}
			if ! reflect.DeepEqual(result, expected) {
				t.Error("Unexpected result in TestNewIPv4Attr(). Structures don't match")
			}
		}
	}
}