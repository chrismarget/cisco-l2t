package attribute

import (
	"bytes"
	"fmt"
	"testing"
)

func TestSpeedAttribute_String(t *testing.T) {
	var (
		speedStringTestData = map[string][]byte{
			"Auto":    []byte{0, 0, 0, 0},
			"10Mb/s":  []byte{0, 0, 0, 1},
			"100Mb/s": []byte{0, 0, 0, 2},
			"1Gb/s":   []byte{0, 0, 0, 3},
			"10Gb/s":  []byte{0, 0, 0, 4},
			"100Gb/s": []byte{0, 0, 0, 5},
			"1Tb/s":   []byte{0, 0, 0, 6},
			"10Tb/s":  []byte{0, 0, 0, 7},
			"100Tb/s": []byte{0, 0, 0, 8},
		}
	)

	for _, speedAttrType := range getAttrsByCategory(speedCategory) {
		for expected, data := range speedStringTestData {
			testAttr := speedAttribute{
				attrType: speedAttrType,
				attrData: data,
			}
			result := testAttr.String()
			if result != expected {
				t.Fatalf("expected %s, got %s", expected, result)
			}
		}
	}
}

func TestSpeedAttribute_Validate_WithGoodData(t *testing.T) {
	goodData := [][]byte{
		[]byte{0, 0, 0, 0},
		[]byte{0, 0, 0, 1},
		[]byte{0, 0, 0, 2},
		[]byte{0, 0, 0, 3},
		[]byte{0, 0, 0, 4},
		[]byte{0, 0, 0, 5},
		[]byte{0, 0, 0, 6},
		[]byte{0, 0, 0, 7},
		[]byte{0, 0, 0, 8},
	}

	for _, speedAttrType := range getAttrsByCategory(speedCategory) {
		for _, testData := range goodData {
			testAttr := speedAttribute{
				attrType: speedAttrType,
				attrData: testData,
			}
			err := testAttr.Validate()
			if err != nil {
				t.Fatalf(err.Error()+"\n"+"Supposed good data %s produced error for %s.",
					fmt.Sprintf("%v", []byte(testAttr.attrData)), attrTypeString[speedAttrType])
			}
		}
	}
}

func TestSpeedAttribute_Validate_WithBadData(t *testing.T) {
	goodData := [][]byte{
		nil,
		[]byte{},
		[]byte{0, 0},
		[]byte{0, 0, 0},
		[]byte{0, 0, 0, 9},
		[]byte{0, 0, 0, 50},
		[]byte{0, 0, 1, 0},
		[]byte{0, 1, 0, 0},
		[]byte{1, 0, 0, 0},
		[]byte{255, 255, 255, 255},
		[]byte{0, 0, 0, 0, 0},
	}

	for _, speedAttrType := range getAttrsByCategory(speedCategory) {
		for _, testData := range goodData {
			testAttr := speedAttribute{
				attrType: speedAttrType,
				attrData: testData,
			}

			err := testAttr.Validate()
			if err == nil {
				t.Fatalf("Bad data %s in %s did not error.",
					fmt.Sprintf("%v", []byte(testAttr.attrData)), attrTypeString[speedAttrType])
			}
		}
	}
}

func TestNewAttrBuilder_Speed_String(t *testing.T) {
	testData := map[string]([]byte) {
		"auto":			[]byte{0, 0, 0, 0},
		"Auto":			[]byte{0, 0, 0, 0},
		"AUTO":			[]byte{0, 0, 0, 0},
		"mb":			[]byte{0, 0, 0, 0},
		"gb/s":			[]byte{0, 0, 0, 0},

		"10mb":			[]byte{0, 0, 0, 1},
		"10mbs":			[]byte{0, 0, 0, 1},
		"10mbps":			[]byte{0, 0, 0, 1},
		"10mb/s":			[]byte{0, 0, 0, 1},

		"100mb":			[]byte{0, 0, 0, 2},
		"100mbs":			[]byte{0, 0, 0, 2},
		"100mbps":			[]byte{0, 0, 0, 2},
		"100mb/s":			[]byte{0, 0, 0, 2},

		"1000mb":			[]byte{0, 0, 0, 3},
		"1000mbs":			[]byte{0, 0, 0, 3},
		"1000mbps":			[]byte{0, 0, 0, 3},
		"1000mb/s":			[]byte{0, 0, 0, 3},
		"1gb":			[]byte{0, 0, 0, 3},
		"1gbs":			[]byte{0, 0, 0, 3},
		"1gbps":			[]byte{0, 0, 0, 3},
		"1gb/s":			[]byte{0, 0, 0, 3},

		"2500mb":			[]byte{0, 0, 0, 3},
		"2500mbs":			[]byte{0, 0, 0, 3},
		"2500mbps":			[]byte{0, 0, 0, 3},
		"2500mb/s":			[]byte{0, 0, 0, 3},
		"2.5gb":			[]byte{0, 0, 0, 3},
		"2.5gbs":			[]byte{0, 0, 0, 3},
		"2.5gbps":			[]byte{0, 0, 0, 3},
		"2.5gb/s":			[]byte{0, 0, 0, 3},

		"5000mb":			[]byte{0, 0, 0, 3},
		"5000mbs":			[]byte{0, 0, 0, 3},
		"5000mbps":			[]byte{0, 0, 0, 3},
		"5000mb/s":			[]byte{0, 0, 0, 3},
		"5gb":			[]byte{0, 0, 0, 3},
		"5gbs":			[]byte{0, 0, 0, 3},
		"5gbps":			[]byte{0, 0, 0, 3},
		"5gb/s":			[]byte{0, 0, 0, 3},

		"10000mb":			[]byte{0, 0, 0, 4},
		"10000mbs":			[]byte{0, 0, 0, 4},
		"10000mbps":			[]byte{0, 0, 0, 4},
		"10000mb/s":			[]byte{0, 0, 0, 4},
		"10gb":			[]byte{0, 0, 0, 4},
		"10gbs":			[]byte{0, 0, 0, 4},
		"10gbps":			[]byte{0, 0, 0, 4},
		"10gb/s":			[]byte{0, 0, 0, 4},

		"25000mb":			[]byte{0, 0, 0, 4},
		"25000mbs":			[]byte{0, 0, 0, 4},
		"25000mbps":			[]byte{0, 0, 0, 4},
		"25000mb/s":			[]byte{0, 0, 0, 4},
		"25gb":			[]byte{0, 0, 0, 4},
		"25gbs":			[]byte{0, 0, 0, 4},
		"25gbps":			[]byte{0, 0, 0, 4},
		"25gb/s":			[]byte{0, 0, 0, 4},

		"50000mb":			[]byte{0, 0, 0, 4},
		"50000mbs":			[]byte{0, 0, 0, 4},
		"50000mbps":			[]byte{0, 0, 0, 4},
		"50000mb/s":			[]byte{0, 0, 0, 4},
		"50gb":			[]byte{0, 0, 0, 4},
		"50gbs":			[]byte{0, 0, 0, 4},
		"50gbps":			[]byte{0, 0, 0, 4},
		"50gb/s":			[]byte{0, 0, 0, 4},

		"100000mb":			[]byte{0, 0, 0, 5},
		"100000mbs":			[]byte{0, 0, 0, 5},
		"100000mbps":			[]byte{0, 0, 0, 5},
		"100000mb/s":			[]byte{0, 0, 0, 5},
		"100gb":			[]byte{0, 0, 0, 5},
		"100gbs":			[]byte{0, 0, 0, 5},
		"100gbps":			[]byte{0, 0, 0, 5},
		"100gb/s":			[]byte{0, 0, 0, 5},

		"200000mb":			[]byte{0, 0, 0, 5},
		"200000mbs":			[]byte{0, 0, 0, 5},
		"200000mbps":			[]byte{0, 0, 0, 5},
		"200000mb/s":			[]byte{0, 0, 0, 5},
		"200gb":			[]byte{0, 0, 0, 5},
		"200gbs":			[]byte{0, 0, 0, 5},
		"200gbps":			[]byte{0, 0, 0, 5},
		"200gb/s":			[]byte{0, 0, 0, 5},

		"400000mb":			[]byte{0, 0, 0, 5},
		"400000mbs":			[]byte{0, 0, 0, 5},
		"400000mbps":			[]byte{0, 0, 0, 5},
		"400000mb/s":			[]byte{0, 0, 0, 5},
		"400gb":			[]byte{0, 0, 0, 5},
		"400gbs":			[]byte{0, 0, 0, 5},
		"400gbps":			[]byte{0, 0, 0, 5},
		"400gb/s":			[]byte{0, 0, 0, 5},
	}

	for _, speedAttrType := range getAttrsByCategory(speedCategory) {
		for s, b := range testData{
			testAttr, err := NewAttrBuilder().SetType(speedAttrType).SetString(s).Build()
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Compare(testAttr.Bytes(), b) != 0 {
				t.Fatalf("speed attribute failed Bytes Comparison: %s vs %s", testAttr.Bytes(), b)
			}
			expected := append([]byte{byte(speedAttrType), 6,}, b...)
			if bytes.Compare(MarshalAttribute(testAttr), expected) != 0 {
				t.Fatalf("speed attribute failed Marshal Comparison: %s vs %s", MarshalAttribute(testAttr), expected)
			}
		}
	}
}

func TestNewAttrBuilder_Speed_Int(t *testing.T) {
	testData := map[uint32]([]byte){
		0: []byte{0, 0, 0, 0},
		10: []byte{0, 0, 0, 1},
		100: []byte{0, 0, 0, 2},
		1000: []byte{0, 0, 0, 3},
		2500: []byte{0, 0, 0, 3},
		5000: []byte{0, 0, 0, 3},
		10000: []byte{0, 0, 0, 4},
		25000: []byte{0, 0, 0, 4},
		50000: []byte{0, 0, 0, 4},
		100000: []byte{0, 0, 0, 5},
		400000: []byte{0, 0, 0, 5},
	}

	for _, speedAttrType := range getAttrsByCategory(speedCategory) {
		for i, b := range testData{
			testAttr, err := NewAttrBuilder().SetType(speedAttrType).SetInt(i).Build()
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Compare(testAttr.Bytes(), b) != 0 {
				t.Fatalf("speed attribute failed Bytes Comparison: %s vs %s", testAttr.Bytes(), b)
			}
			expected := append([]byte{byte(speedAttrType), 6,}, b...)
			if bytes.Compare(MarshalAttribute(testAttr), expected) != 0 {
				t.Fatalf("speed attribute failed Marshal Comparison: %s vs %s", MarshalAttribute(testAttr), expected)
			}
		}
	}
}

func TestNewAttrBuilder_Speed_Bytes(t *testing.T) {
	testData := [][]byte{
		[]byte{0, 0, 0, 0},
		[]byte{0, 0, 0, 1},
		[]byte{0, 0, 0, 2},
		[]byte{0, 0, 0, 3},
		[]byte{0, 0, 0, 4},
		[]byte{0, 0, 0, 5},
		[]byte{0, 0, 0, 6},
		[]byte{0, 0, 0, 7},
		[]byte{0, 0, 0, 8},
	}

	for _, speedAttrType := range getAttrsByCategory(speedCategory) {
		for _, b := range testData{
			testAttr, err := NewAttrBuilder().SetType(speedAttrType).SetBytes(b).Build()
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Compare(testAttr.Bytes(), b) != 0 {
				t.Fatalf("speed attribute failed Bytes Comparison: %s vs %s", testAttr.Bytes(), b)
			}
			expected := append([]byte{byte(speedAttrType), 6,}, b...)
			if bytes.Compare(MarshalAttribute(testAttr), expected) != 0 {
				t.Fatalf("speed attribute failed Marshal Comparison: %s vs %s", MarshalAttribute(testAttr), expected)
			}
		}
	}
}

//func TestNewSpeedAttrWithInt(t *testing.T) {
//	attrTypesToTest := getAttrsByCategory(speedCategory)
//	for _, testType := range attrTypesToTest {
//		var intToTest []int
//		intToTest = append(intToTest, 0)
//		intToTest = append(intToTest, 1)
//		intToTest = append(intToTest, 2)
//		intToTest = append(intToTest, 3)
//		intToTest = append(intToTest, 4)
//		intToTest = append(intToTest, 5)
//		intToTest = append(intToTest, 10)
//		intToTest = append(intToTest, 100)
//		intToTest = append(intToTest, 1000)
//		intToTest = append(intToTest, 2500)
//		intToTest = append(intToTest, 5000)
//		intToTest = append(intToTest, 10000)
//		intToTest = append(intToTest, 25000)
//		intToTest = append(intToTest, 50000)
//		intToTest = append(intToTest, 100000)
//		intToTest = append(intToTest, 200000)
//		intToTest = append(intToTest, 400000)
//
//		var expected []Attr
//		expected = append(expected, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 0}})
//		expected = append(expected, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 1}})
//		expected = append(expected, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 2}})
//		expected = append(expected, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 3}})
//		expected = append(expected, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 4}})
//		expected = append(expected, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 5}})
//		expected = append(expected, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 1}})
//		expected = append(expected, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 2}})
//		expected = append(expected, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 3}})
//		expected = append(expected, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 3}})
//		expected = append(expected, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 3}})
//		expected = append(expected, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 4}})
//		expected = append(expected, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 4}})
//		expected = append(expected, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 4}})
//		expected = append(expected, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 5}})
//		expected = append(expected, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 5}})
//		expected = append(expected, Attr{AttrType: testType, AttrData: []byte{0, 0, 0, 5}})
//
//		for k, _ := range intToTest {
//			result, err := NewAttr(testType, attrPayload{intData: intToTest[k]})
//			if err != nil {
//				t.Error(err)
//			}
//			if !reflect.DeepEqual(result, expected[k]) {
//				t.Error("Structures don't match")
//			}
//		}
//	}
//}
//
//func TestNewSpeedAttrWithString(t *testing.T) {
//	attrTypesToTest := getAttrsByCategory(speedCategory)
//	for _, testType := range attrTypesToTest {
//		var stringsToTest []string
//		stringsToTest = append(stringsToTest, "auto")
//		stringsToTest = append(stringsToTest, "Auto")
//		stringsToTest = append(stringsToTest, "AUTO")
//		stringsToTest = append(stringsToTest, "mb")
//		stringsToTest = append(stringsToTest, "gb/s")
//
//		stringsToTest = append(stringsToTest, "10mb")
//		stringsToTest = append(stringsToTest, "10mbs")
//		stringsToTest = append(stringsToTest, "10mbps")
//		stringsToTest = append(stringsToTest, "10mb/s")
//
//		stringsToTest = append(stringsToTest, "100mb")
//		stringsToTest = append(stringsToTest, "100mbs")
//		stringsToTest = append(stringsToTest, "100mbps")
//		stringsToTest = append(stringsToTest, "100mb/s")
//
//		stringsToTest = append(stringsToTest, "1000mb")
//		stringsToTest = append(stringsToTest, "1000mbs")
//		stringsToTest = append(stringsToTest, "1000mbps")
//		stringsToTest = append(stringsToTest, "1000mb/s")
//		stringsToTest = append(stringsToTest, "1gb")
//		stringsToTest = append(stringsToTest, "1gbs")
//		stringsToTest = append(stringsToTest, "1gbps")
//		stringsToTest = append(stringsToTest, "1gb/s")
//
//		stringsToTest = append(stringsToTest, "2500mb")
//		stringsToTest = append(stringsToTest, "2500mbs")
//		stringsToTest = append(stringsToTest, "2500mbps")
//		stringsToTest = append(stringsToTest, "2500mb/s")
//		stringsToTest = append(stringsToTest, "2.5gb")
//		stringsToTest = append(stringsToTest, "2.5gbs")
//		stringsToTest = append(stringsToTest, "2.5gbps")
//		stringsToTest = append(stringsToTest, "2.5gb/s")
//
//		stringsToTest = append(stringsToTest, "5000mb")
//		stringsToTest = append(stringsToTest, "5000mbs")
//		stringsToTest = append(stringsToTest, "5000mbps")
//		stringsToTest = append(stringsToTest, "5000mb/s")
//		stringsToTest = append(stringsToTest, "5gb")
//		stringsToTest = append(stringsToTest, "5gbs")
//		stringsToTest = append(stringsToTest, "5gbps")
//		stringsToTest = append(stringsToTest, "5gb/s")
//
//		stringsToTest = append(stringsToTest, "10000mb")
//		stringsToTest = append(stringsToTest, "10000mbs")
//		stringsToTest = append(stringsToTest, "10000mbps")
//		stringsToTest = append(stringsToTest, "10000mb/s")
//		stringsToTest = append(stringsToTest, "10gb")
//		stringsToTest = append(stringsToTest, "10gbs")
//		stringsToTest = append(stringsToTest, "10gbps")
//		stringsToTest = append(stringsToTest, "10gb/s")
//
//		stringsToTest = append(stringsToTest, "25000mb")
//		stringsToTest = append(stringsToTest, "25000mbs")
//		stringsToTest = append(stringsToTest, "25000mbps")
//		stringsToTest = append(stringsToTest, "25000mb/s")
//		stringsToTest = append(stringsToTest, "25gb")
//		stringsToTest = append(stringsToTest, "25gbs")
//		stringsToTest = append(stringsToTest, "25gbps")
//		stringsToTest = append(stringsToTest, "25gb/s")
//
//		stringsToTest = append(stringsToTest, "50000mb")
//		stringsToTest = append(stringsToTest, "50000mbs")
//		stringsToTest = append(stringsToTest, "50000mbps")
//		stringsToTest = append(stringsToTest, "50000mb/s")
//		stringsToTest = append(stringsToTest, "50gb")
//		stringsToTest = append(stringsToTest, "50gbs")
//		stringsToTest = append(stringsToTest, "50gbps")
//		stringsToTest = append(stringsToTest, "50gb/s")
//
//		stringsToTest = append(stringsToTest, "100000mb")
//		stringsToTest = append(stringsToTest, "100000mbs")
//		stringsToTest = append(stringsToTest, "100000mbps")
//		stringsToTest = append(stringsToTest, "100000mb/s")
//		stringsToTest = append(stringsToTest, "100gb")
//		stringsToTest = append(stringsToTest, "100gbs")
//		stringsToTest = append(stringsToTest, "100gbps")
//		stringsToTest = append(stringsToTest, "100gb/s")
//
//		stringsToTest = append(stringsToTest, "200000mb")
//		stringsToTest = append(stringsToTest, "200000mbs")
//		stringsToTest = append(stringsToTest, "200000mbps")
//		stringsToTest = append(stringsToTest, "200000mb/s")
//		stringsToTest = append(stringsToTest, "200gb")
//		stringsToTest = append(stringsToTest, "200gbs")
//		stringsToTest = append(stringsToTest, "200gbps")
//		stringsToTest = append(stringsToTest, "200gb/s")
//
//		stringsToTest = append(stringsToTest, "400000mb")
//		stringsToTest = append(stringsToTest, "400000mbs")
//		stringsToTest = append(stringsToTest, "400000mbps")
//		stringsToTest = append(stringsToTest, "400000mb/s")
//		stringsToTest = append(stringsToTest, "400gb")
//		stringsToTest = append(stringsToTest, "400gbs")
//		stringsToTest = append(stringsToTest, "400gbps")
//		stringsToTest = append(stringsToTest, "400gb/s")
//
//
//		for k, testString := range stringsToTest {
//			result, err := NewAttr(testType, attrPayload{stringData: testString})
//			if err != nil {
//				t.Error(err)
//			}
//			expected := expectedResults[k]
//			if !reflect.DeepEqual(result, expected) {
//				t.Error("Structures don't match.")
//			}
//		}
//	}
//}
//
//func TestValidateSpeed(t *testing.T) {
//	speedTypes := map[attrType]bool{}
//	for i := 0; i <= 255; i++ {
//		if attrCategoryByType[attrType(i)] == speedCategory {
//			speedTypes[attrType(i)] = true
//		}
//	}
//}
