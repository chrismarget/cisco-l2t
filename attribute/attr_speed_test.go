package attribute

import (
	"encoding/binary"
	"reflect"
	"testing"
)

func TestStringSpeed(t *testing.T) {
	attrTypesToTest := getAttrsByCategory(speedCategory)
	for _, v := range attrTypesToTest {
		speedVals := []uint32{0, 1, 2, 3, 4, 5}
		expectedVals := []string{
			"Auto",
			"10Mb/s",
			"100Mb/s",
			"1Gb/s",
			"10Gb/s",
			"100Gb/s",
		}
		for k, _ := range speedVals {
			d := make([]byte, 4)
			binary.BigEndian.PutUint32(d, speedVals[k])
			data := Attr{
				AttrType: v,
				AttrData: d,
			}
			result, err := data.String()
			if err != nil {
				t.Error(err)
			}
			expected := expectedVals[k]
			if result != expected {
				t.Errorf("expected '%s', got '%s'", expected, result)
			}

		}
	}
}

func TestNewSpeedAttrWithInt(t *testing.T) {
	attrTypesToTest := getAttrsByCategory(speedCategory)
	for _, testType := range attrTypesToTest {
		var intToTest []int
		intToTest = append(intToTest, 0)
		intToTest = append(intToTest, 1)
		intToTest = append(intToTest, 2)
		intToTest = append(intToTest, 3)
		intToTest = append(intToTest, 4)
		intToTest = append(intToTest, 5)
		intToTest = append(intToTest, 10)
		intToTest = append(intToTest, 100)
		intToTest = append(intToTest, 1000)
		intToTest = append(intToTest, 2500)
		intToTest = append(intToTest, 5000)
		intToTest = append(intToTest, 10000)
		intToTest = append(intToTest, 25000)
		intToTest = append(intToTest, 50000)
		intToTest = append(intToTest, 100000)
		intToTest = append(intToTest, 200000)
		intToTest = append(intToTest, 400000)

		var expected []Attr
		expected = append(expected, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 0}})
		expected = append(expected, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 1}})
		expected = append(expected, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 2}})
		expected = append(expected, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 3}})
		expected = append(expected, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 4}})
		expected = append(expected, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 5}})
		expected = append(expected, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 1}})
		expected = append(expected, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 2}})
		expected = append(expected, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 3}})
		expected = append(expected, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 3}})
		expected = append(expected, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 3}})
		expected = append(expected, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 4}})
		expected = append(expected, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 4}})
		expected = append(expected, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 4}})
		expected = append(expected, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 5}})
		expected = append(expected, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 5}})
		expected = append(expected, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 5}})

		for k, _ := range intToTest {
			result, err := NewAttr(testType, attrPayload{intData: intToTest[k]})
			if err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(result, expected[k]) {
				t.Error("Structures don't match")
			}
		}
	}
}

func TestNewSpeedAttrWithString(t *testing.T) {
	attrTypesToTest := getAttrsByCategory(speedCategory)
	for _, testType := range attrTypesToTest {
		var stringsToTest []string
		stringsToTest = append(stringsToTest, "auto")
		stringsToTest = append(stringsToTest, "Auto")
		stringsToTest = append(stringsToTest, "AUTO")
		stringsToTest = append(stringsToTest, "mb")
		stringsToTest = append(stringsToTest, "gb/s")

		stringsToTest = append(stringsToTest, "10mb")
		stringsToTest = append(stringsToTest, "10mbs")
		stringsToTest = append(stringsToTest, "10mbps")
		stringsToTest = append(stringsToTest, "10mb/s")

		stringsToTest = append(stringsToTest, "100mb")
		stringsToTest = append(stringsToTest, "100mbs")
		stringsToTest = append(stringsToTest, "100mbps")
		stringsToTest = append(stringsToTest, "100mb/s")

		stringsToTest = append(stringsToTest, "1000mb")
		stringsToTest = append(stringsToTest, "1000mbs")
		stringsToTest = append(stringsToTest, "1000mbps")
		stringsToTest = append(stringsToTest, "1000mb/s")
		stringsToTest = append(stringsToTest, "1gb")
		stringsToTest = append(stringsToTest, "1gbs")
		stringsToTest = append(stringsToTest, "1gbps")
		stringsToTest = append(stringsToTest, "1gb/s")

		stringsToTest = append(stringsToTest, "2500mb")
		stringsToTest = append(stringsToTest, "2500mbs")
		stringsToTest = append(stringsToTest, "2500mbps")
		stringsToTest = append(stringsToTest, "2500mb/s")
		stringsToTest = append(stringsToTest, "2.5gb")
		stringsToTest = append(stringsToTest, "2.5gbs")
		stringsToTest = append(stringsToTest, "2.5gbps")
		stringsToTest = append(stringsToTest, "2.5gb/s")

		stringsToTest = append(stringsToTest, "5000mb")
		stringsToTest = append(stringsToTest, "5000mbs")
		stringsToTest = append(stringsToTest, "5000mbps")
		stringsToTest = append(stringsToTest, "5000mb/s")
		stringsToTest = append(stringsToTest, "5gb")
		stringsToTest = append(stringsToTest, "5gbs")
		stringsToTest = append(stringsToTest, "5gbps")
		stringsToTest = append(stringsToTest, "5gb/s")

		stringsToTest = append(stringsToTest, "10000mb")
		stringsToTest = append(stringsToTest, "10000mbs")
		stringsToTest = append(stringsToTest, "10000mbps")
		stringsToTest = append(stringsToTest, "10000mb/s")
		stringsToTest = append(stringsToTest, "10gb")
		stringsToTest = append(stringsToTest, "10gbs")
		stringsToTest = append(stringsToTest, "10gbps")
		stringsToTest = append(stringsToTest, "10gb/s")

		stringsToTest = append(stringsToTest, "25000mb")
		stringsToTest = append(stringsToTest, "25000mbs")
		stringsToTest = append(stringsToTest, "25000mbps")
		stringsToTest = append(stringsToTest, "25000mb/s")
		stringsToTest = append(stringsToTest, "25gb")
		stringsToTest = append(stringsToTest, "25gbs")
		stringsToTest = append(stringsToTest, "25gbps")
		stringsToTest = append(stringsToTest, "25gb/s")

		stringsToTest = append(stringsToTest, "50000mb")
		stringsToTest = append(stringsToTest, "50000mbs")
		stringsToTest = append(stringsToTest, "50000mbps")
		stringsToTest = append(stringsToTest, "50000mb/s")
		stringsToTest = append(stringsToTest, "50gb")
		stringsToTest = append(stringsToTest, "50gbs")
		stringsToTest = append(stringsToTest, "50gbps")
		stringsToTest = append(stringsToTest, "50gb/s")

		stringsToTest = append(stringsToTest, "100000mb")
		stringsToTest = append(stringsToTest, "100000mbs")
		stringsToTest = append(stringsToTest, "100000mbps")
		stringsToTest = append(stringsToTest, "100000mb/s")
		stringsToTest = append(stringsToTest, "100gb")
		stringsToTest = append(stringsToTest, "100gbs")
		stringsToTest = append(stringsToTest, "100gbps")
		stringsToTest = append(stringsToTest, "100gb/s")

		stringsToTest = append(stringsToTest, "200000mb")
		stringsToTest = append(stringsToTest, "200000mbs")
		stringsToTest = append(stringsToTest, "200000mbps")
		stringsToTest = append(stringsToTest, "200000mb/s")
		stringsToTest = append(stringsToTest, "200gb")
		stringsToTest = append(stringsToTest, "200gbs")
		stringsToTest = append(stringsToTest, "200gbps")
		stringsToTest = append(stringsToTest, "200gb/s")

		stringsToTest = append(stringsToTest, "400000mb")
		stringsToTest = append(stringsToTest, "400000mbs")
		stringsToTest = append(stringsToTest, "400000mbps")
		stringsToTest = append(stringsToTest, "400000mb/s")
		stringsToTest = append(stringsToTest, "400gb")
		stringsToTest = append(stringsToTest, "400gbs")
		stringsToTest = append(stringsToTest, "400gbps")
		stringsToTest = append(stringsToTest, "400gb/s")

		var expectedResults []Attr
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 0}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 0}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 0}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 0}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 0}})

		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 1}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 1}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 1}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 1}})

		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 2}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 2}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 2}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 2}})

		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 3}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 3}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 3}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 3}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 3}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 3}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 3}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 3}})

		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 3}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 3}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 3}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 3}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 3}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 3}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 3}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 3}})

		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 3}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 3}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 3}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 3}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 3}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 3}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 3}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 3}})

		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 4}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 4}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 4}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 4}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 4}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 4}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 4}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 4}})

		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 4}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 4}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 4}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 4}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 4}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 4}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 4}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 4}})

		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 4}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 4}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 4}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 4}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 4}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 4}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 4}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 4}})

		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 5}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 5}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 5}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 5}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 5}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 5}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 5}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 5}})

		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 5}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 5}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 5}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 5}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 5}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 5}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 5}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 5}})

		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 5}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 5}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 5}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 5}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 5}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 5}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 5}})
		expectedResults = append(expectedResults, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 5}})

		for k, testString := range stringsToTest {
			result, err := NewAttr(testType, attrPayload{stringData: testString})
			if err != nil {
				t.Error(err)
			}
			expected := expectedResults[k]
			if !reflect.DeepEqual(result, expected) {
				t.Error("Structures don't match.")
			}
		}
	}
}

func TestValidateSpeed(t *testing.T) {
	speedTypes := map[attrType]bool{}
	for i := 0; i <= 255; i++ {
		if attrCategoryByType[attrType(i)] == speedCategory {
			speedTypes[attrType(i)] = true
		}
	}
}
