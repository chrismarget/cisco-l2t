package attribute

import (
	"fmt"
	"math"
	"testing"
)

func TestDuplexAttribute_String(t *testing.T) {
	var (
		duplexStringTestData = map[string]portDuplex{
			"Auto": portDuplex(0),
			"Half": portDuplex(1),
			"Full": portDuplex(2),
		}
	)

	for _, duplexAttrType := range  getAttrsByCategory(duplexCategory) {
		for expected, data := range duplexStringTestData {
			testAttr := duplexAttribute{
				attrType: duplexAttrType,
				attrData: []byte{byte(data)},
			}
			result := testAttr.String()
			if result != expected {
				t.Fatalf("expected %s, got %s", expected, result)
			}
		}
	}
}

func TestDuplexAttribute_Validate_WithGoodData(t *testing.T) {
	goodData := [][]byte {
		[]byte{0},
		[]byte{1},
		[]byte{2},
	}
	for _, duplexAttrType := range  getAttrsByCategory(duplexCategory) {
		for _, testData := range goodData {
			testAttr := duplexAttribute{
				attrType: duplexAttrType,
				attrData: testData,
			}
			err := testAttr.Validate()
			if err != nil {
				t.Fatalf(err.Error()+"\n"+"Supposed good data %s produced error for %s.",
					fmt.Sprintf("%v", []byte(testAttr.attrData)), attrTypeString[duplexAttrType])
			}
		}
	}
}

func TestDuplexAttribute_Validate_WithBadData(t *testing.T) {
	badData := [][]byte {
		nil,
		[]byte{},
		[]byte {0,0},
	}

	for i := 3; i <= math.MaxUint8; i++ {
		badData = append(badData, []byte{byte(i)})
	}

	for _, duplexAttrType := range  getAttrsByCategory(duplexCategory) {
		for _, testData := range badData {
			testAttr := duplexAttribute{
				attrType: duplexAttrType,
				attrData: testData,
			}

			err := testAttr.Validate()
			if err == nil {
				t.Fatalf("Bad data %s in %s did not error.",
					fmt.Sprintf("%v", []byte(testAttr.attrData)), attrTypeString[duplexAttrType])
			}
		}
	}
}

