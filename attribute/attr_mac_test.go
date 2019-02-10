package attribute

import (
	"fmt"
	"testing"
)

func TestMacAttribute_String(t *testing.T) {
	var (
		macStringTestData = map[string][]byte{
			"00:00:00:00:00:00": []byte{0,0,0,0,0,0},
			"01:02:03:04:05:06": []byte{1,2,3,4,5,6},
			"ff:ff:ff:ff:ff:ff": []byte{255,255,255,255,255,255},
		}
	)

	for _, macAttrType := range  getAttrsByCategory(macCategory) {
		for expected, data := range macStringTestData {
			testAttr := macAttribute{
				attrType: macAttrType,
				attrData: data,
			}
			result := testAttr.String()
			if result != expected {
				t.Fatalf("expected %s, got %s", expected, result)
			}
		}
	}
}

func TestMacAttribute_Validate_WithGoodData(t *testing.T) {
	goodData := [][]byte {
		[]byte{0,0,0,0,0,0},
		[]byte{1,2,3,4,5,6},
		[]byte{255,255,255,255,255,255},
	}

	for _, macAttrType := range  getAttrsByCategory(macCategory) {
		for _, testData := range goodData {
			testAttr := macAttribute{
				attrType: macAttrType,
				attrData: testData,
			}
			err := testAttr.Validate()
			if err != nil {
				t.Fatalf(err.Error()+"\n"+"Supposed good data %s produced error for %s.",
					fmt.Sprintf("%v", []byte(testAttr.attrData)), attrTypeString[macAttrType])
			}
		}
	}
}

func TestMacAttribute_Validate_WithBadData(t *testing.T) {
	goodData := [][]byte {
		nil,
		[]byte{},
		[]byte{0},
		[]byte{0,0},
		[]byte{0,0,0},
		[]byte{0,0,0,0},
		[]byte{0,0,0,0,0},
		[]byte{0,0,0,0,0,0,0},
	}

	for _, macAttrType := range  getAttrsByCategory(macCategory) {
		for _, testData := range goodData {
			testAttr := macAttribute{
				attrType: macAttrType,
				attrData: testData,
			}

			err := testAttr.Validate()
			if err == nil {
				t.Fatalf("Bad data %s in %s did not error.",
					fmt.Sprintf("%v", []byte(testAttr.attrData)), attrTypeString[macAttrType])
			}
		}
	}
}
