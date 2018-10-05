package attribute

import (
	"net"
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

		data2 := attr{
			attrType: v,
			attrData: []byte{192, 168, 10, 11, 12},
		}
		_, err = data2.String()
		if err == nil {
			t.Error("Oversize IP payload should have produced an error")
		}

		data3 := attr{
			attrType: v,
			attrData: []byte{192, 168, 10},
		}
		_, err = data3.String()
		if err == nil {
			t.Error("Undersize IP payload should have produced an error")
		}
	}
}

func TestNewIPv4AttrStringPayload(t *testing.T) {
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

func TestNewIPv4AttrIntPayload(t *testing.T) {
	attrTypesToTest := getAttrsByCategory(ipv4Category)
	var testType attrType

	for _, testType = range attrTypesToTest {
		intsToTest := []int{0, 3232238091, 4294967295}
		expectedResults := [][]byte{{0,0,0,0}, {192, 168,10,11}, {255,255,255,255}}
		for k, _ := range intsToTest {
			var testPayload attrPayload
			var expected attr
			var err error
			var result attr

			expected = attr{
				attrType: testType,
				attrData: expectedResults[k],
			}

			testPayload.intData = intsToTest[k]
			result, err = NewAttr(testType, testPayload)
			if err != nil {
				t.Error(err)
			}
			if ! reflect.DeepEqual(result, expected) {
				t.Error("Unexpected result in TestNewIPv4Attr(). Structures don't match")
			}
		}
	}

	for _, testType = range attrTypesToTest {
		intsToTest := []int{-1, 4294967296}

		for k, _ := range intsToTest {
			var testPayload attrPayload
			var err error

			testPayload.intData = intsToTest[k]
			_, err = NewAttr(testType, testPayload)
			if err == nil {
				t.Error("Out of range integers should produce an error")
			}
		}
	}
}

func TestNewIPv4AttrIPAddrPayload(t *testing.T) {
	attrTypesToTest := getAttrsByCategory(ipv4Category)
	var testType attrType

	for _, testType = range attrTypesToTest {
		iPAddrsToTest := []net.IPAddr{
			net.IPAddr{IP: []byte{0,0,0,0}},
			net.IPAddr{IP: []byte{192,168,11,12}},
			net.IPAddr{IP: []byte{255,255,255,255}},
		}
		expectedAttrData := [][]byte{{0, 0, 0, 0}, {192, 168, 11, 12}, {255, 255, 255, 255}}
		for k, _ := range iPAddrsToTest {
			var testPayload attrPayload
			var expected attr
			var err error
			var result attr

			testPayload.ipAddrData = iPAddrsToTest[k]
			result, err = NewAttr(testType, testPayload)
			if err != nil {
				t.Error(err)
			}

			expected = attr{attrType: testType, attrData: expectedAttrData[k]}
			if ! reflect.DeepEqual(result, expected) {
				t.Error("Unexpected result in TestNewIPv4Attr(). Structures don't match")
			}

		}
	}
}
