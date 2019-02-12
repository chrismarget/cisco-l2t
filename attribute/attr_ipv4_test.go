package attribute

import (
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestStringIPv4(t *testing.T) {
	attrTypesToTest := getAttrsByCategory(ipv4Category)
	for _, v := range attrTypesToTest {
		data1 := Attr{
			AttrType: v,
			AttrData: []byte{192, 168, 10, 11},
		}
		expected1 := "192.168.10.11"
		result1, err := data1.String()
		if err != nil {
			t.Error(err)
		}
		if result1 != expected1 {
			t.Errorf("expected '%s', got '%s'", expected1, result1)
		}

		data2 := Attr{
			AttrType: v,
			AttrData: []byte{192, 168, 10, 11, 12},
		}
		_, err = data2.String()
		if err == nil {
			t.Error("Oversize IP payload should have produced an error")
		}

		data3 := Attr{
			AttrType: v,
			AttrData: []byte{192, 168, 10},
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
			var expected Attr
			var err error
			var result Attr

			// building the "expected" structure requires us to walk the test
			// string the hard way.
			octets := strings.Split(testString, ".")
			var data []byte
			for _, o := range octets {
				var i int
				i, err = strconv.Atoi(o)
				data = append(data, byte(i))
			}
			expected = Attr{
				AttrType: testType,
				AttrData: data,
			}

			testPayload.stringData = testString
			result, err = NewAttr(testType, testPayload)
			if err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(result, expected) {
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
		expectedResults := [][]byte{{0, 0, 0, 0}, {192, 168, 10, 11}, {255, 255, 255, 255}}
		for k, _ := range intsToTest {
			var testPayload attrPayload
			var expected Attr
			var err error
			var result Attr

			expected = Attr{
				AttrType: testType,
				AttrData: expectedResults[k],
			}

			testPayload.intData = intsToTest[k]
			result, err = NewAttr(testType, testPayload)
			if err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(result, expected) {
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
			net.IPAddr{IP: []byte{0, 0, 0, 0}},
			net.IPAddr{IP: []byte{192, 168, 11, 12}},
			net.IPAddr{IP: []byte{255, 255, 255, 255}},
		}
		expectedAttrData := [][]byte{{0, 0, 0, 0}, {192, 168, 11, 12}, {255, 255, 255, 255}}
		for k, _ := range iPAddrsToTest {
			var testPayload attrPayload
			var expected Attr
			var err error
			var result Attr

			testPayload.ipAddrData = iPAddrsToTest[k]
			result, err = NewAttr(testType, testPayload)
			if err != nil {
				t.Error(err)
			}

			expected = Attr{AttrType: testType, AttrData: expectedAttrData[k]}
			if !reflect.DeepEqual(result, expected) {
				t.Error("Unexpected result in TestNewIPv4Attr(). Structures don't match")
			}

		}
	}
}

func TestValidateIPv4(t *testing.T) {
	iPv4Types := map[attrType]bool{}
	for i := 0; i <= 255; i++ {
		if attrCategoryByType[attrType(i)] == ipv4Category {
			iPv4Types[attrType(i)] = true
		}
	}

	for i := 0; i <= 255; i++ {
		a := Attr{AttrType: attrType(i), AttrData: []byte{0, 0, 0, 0}}
		err := validateIPv4(a)
		switch {
		case iPv4Types[attrType(i)] && err != nil:
			msg := fmt.Sprintf("Attribute type %d should not produce IPv4 validation errors: %s", i, err)
			t.Error(msg)
		case !iPv4Types[attrType(i)] && err == nil:
			msg := fmt.Sprintf("Attribute type %d should have produced IPv4 validation errors.", i)
			t.Error(msg)
		}
	}
}
