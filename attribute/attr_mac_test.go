package attribute

import (
	"reflect"
	"testing"
)

func TestStringMac(t *testing.T) {
	attrTypesToTest := getAttrsByCategory(macCategory)
	for _, testType := range attrTypesToTest {
		data1 := attr{
			attrType: testType,
			attrData: []byte{0, 0, 0, 0, 0, 0},
		}
		expected1 := "00:00:00:00:00:00"
		result1, err := data1.String()
		if err != nil {
			t.Error(err)
		}
		if result1 != expected1 {
			t.Errorf("expected '%s', got '%s'", expected1, result1)
		}

		data2 := attr{
			attrType: testType,
			attrData: []byte{255, 255, 255, 255, 255, 255},
		}
		expected2 := "ff:ff:ff:ff:ff:ff"
		result2, err := data2.String()
		if err != nil {
			t.Error(err)
		}
		if result2 != expected2 {
			t.Errorf("expected '%s', got '%s'", expected2, result2)
		}

		data3 := attr{
			attrType: testType,
			attrData: []byte{0, 0, 0, 0, 0},
		}
		_, err = data3.String()
		if err == nil {
			t.Error("Undersize MAC payload should have generated and error")
		}

		data4 := attr{
			attrType: testType,
			attrData: []byte{0, 0, 0, 0, 0, 0, 0},
		}
		_, err = data4.String()
		if err == nil {
			t.Error("Oversize MAC payload should have generated and error")
		}
	}
}

func TestNewMacAttrWithString(t *testing.T) {
	attrTypesToTest := getAttrsByCategory(macCategory)
	for _, testType := range attrTypesToTest {
		var stringsToTest []string
		stringsToTest = append(stringsToTest, "00:ff:ff:ff:ff:01")
		stringsToTest = append(stringsToTest, "00:FF:FF:FF:FF:02")
		//stringsToTest = append(stringsToTest, "0:ff:ff:ff:ff:3")
		//stringsToTest = append(stringsToTest, "0:FF:FF:FF:FF:4")
		stringsToTest = append(stringsToTest, "00-ff-ff-ff-ff-05")
		stringsToTest = append(stringsToTest, "00-FF-FF-FF-FF-06")
		//stringsToTest = append(stringsToTest, "0-ff-ff-ff-ff-7")
		//stringsToTest = append(stringsToTest, "0-FF-FF-FF-FF-8")
		stringsToTest = append(stringsToTest, "00ff.ffff.ff09")
		stringsToTest = append(stringsToTest, "00FF.FFFF.FF0A")
		//stringsToTest = append(stringsToTest, "ff.ffff.ff0b")
		//stringsToTest = append(stringsToTest, "FF.FFFF.FF0C")

		var expectedAttrData [][]byte
		expectedAttrData = append(expectedAttrData, []byte{0, 255, 255, 255, 255, 1})
		expectedAttrData = append(expectedAttrData, []byte{0, 255, 255, 255, 255, 2})
		//expectedAttrData = append(expectedAttrData, []byte{0, 255, 255, 255, 255, 3})
		//expectedAttrData = append(expectedAttrData, []byte{0, 255, 255, 255, 255, 4})
		expectedAttrData = append(expectedAttrData, []byte{0, 255, 255, 255, 255, 5})
		expectedAttrData = append(expectedAttrData, []byte{0, 255, 255, 255, 255, 6})
		//expectedAttrData = append(expectedAttrData, []byte{0, 255, 255, 255, 255, 7})
		//expectedAttrData = append(expectedAttrData, []byte{0, 255, 255, 255, 255, 8})
		expectedAttrData = append(expectedAttrData, []byte{0, 255, 255, 255, 255, 9})
		expectedAttrData = append(expectedAttrData, []byte{0, 255, 255, 255, 255, 10})
		//expectedAttrData = append(expectedAttrData, []byte{0, 255, 255, 255, 255, 11})
		//expectedAttrData = append(expectedAttrData, []byte{0, 255, 255, 255, 255, 12})

		var expectedAttr []attr
		for _, v := range expectedAttrData {
			expectedAttr = append(expectedAttr, attr{attrType: testType, attrData: v})
		}

		for k, v := range stringsToTest {
			result, err := NewAttr(testType, attrPayload{stringData: v})
			if err != nil {
				t.Error(err)
			}

			if !reflect.DeepEqual(result, expectedAttr[k]) {
				t.Error("Structures don't match.")
			}
		}
	}

	for _, testType := range attrTypesToTest {
		testString := "bogus"
		_, err := NewAttr(testType, attrPayload{stringData: testString})
		if err == nil {
			t.Error("Error: Bogus string should have produced an error.")
		}
	}
}

func TestNewMacAttrWithInt(t *testing.T) {
	attrTypesToTest := getAttrsByCategory(macCategory)
	for _, testType := range attrTypesToTest {
		var intsToTest []int
		intsToTest = append(intsToTest, 0)
		intsToTest = append(intsToTest, 1)

		var expectedAttrData [][]byte
		expectedAttrData = append(expectedAttrData, []byte{0,0,0,0,0,0})
		expectedAttrData = append(expectedAttrData, []byte{0,0,0,0,0,1})

		var expectedAttrs []attr
		for _, v := range expectedAttrData {
			expectedAttrs = append(expectedAttrs, attr{attrType: testType, attrData: v})
		}

		for k, v := range intsToTest {
			result, err := NewAttr(testType, attrPayload{intData: v})
			if err != nil {
				t.Error(err)
			}

			if ! reflect.DeepEqual(result, expectedAttrs[k]) {
				t.Error("Structures don't match.")
			}
		}
	}
}
