package attribute

import (
	"fmt"
	"testing"
)

func TestIpv4Attribute_String(t *testing.T) {
	var (
		ipv4StringTestData = map[string][]byte{
			"0.0.0.0": []byte{0,0,0,0},
			"1.2.3.4": []byte{1,2,3,4},
			"192.168.34.56": []byte{192,168,34,56},
			"224.1.2.3": []byte{224,1,2,3},
			"255.255.255.255": []byte{255,255,255,255},
		}
	)

	for _, ipv4AttrType := range  getAttrsByCategory(ipv4Category) {
		for expected, data := range ipv4StringTestData {
			testAttr := ipv4Attribute{
				attrType: ipv4AttrType,
				attrData: data,
			}
			result := testAttr.String()
			if result != expected {
				t.Fatalf("expected %s, got %s", expected, result)
			}
		}
	}
}

func TestIpv4Attribute_Validate_WithGoodData(t *testing.T) {
	goodData := [][]byte {
		[]byte{0,0,0,0},
		[]byte{1,2,3,4},
		[]byte{192,168,34,56},
		[]byte{224,1,2,3},
		[]byte{255,255,255,255},
	}

	for _, ipv4AttrType := range  getAttrsByCategory(ipv4Category) {
		for _, testData := range goodData {
			testAttr := ipv4Attribute{
				attrType: ipv4AttrType,
				attrData: testData,
			}
			err := testAttr.Validate()
			if err != nil {
				t.Fatalf(err.Error()+"\n"+"Supposed good data %s produced error for %s.",
					fmt.Sprintf("%v", []byte(testAttr.attrData)), attrTypeString[ipv4AttrType])
			}
		}
	}
}

func TestIpv4Attribute_Validate_WithBadData(t *testing.T) {
	goodData := [][]byte {
		nil,
		[]byte{},
		[]byte{0},
		[]byte{0,0},
		[]byte{0,0,0},
		[]byte{0,0,0,0,0},
	}

	for _, ipv4AttrType := range  getAttrsByCategory(ipv4Category) {
		for _, testData := range goodData {
			testAttr := ipv4Attribute{
				attrType: ipv4AttrType,
				attrData: testData,
			}

			err := testAttr.Validate()
			if err == nil {
					t.Fatalf("Bad data %s in %s did not error.",
					fmt.Sprintf("%v", []byte(testAttr.attrData)), attrTypeString[ipv4AttrType])
			}
		}
	}
}
