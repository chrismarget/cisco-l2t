package attribute

import (
	"bytes"
	"fmt"
	"testing"
)

func TestIpv4Attribute_String(t *testing.T) {
	var (
		ipv4StringTestData = map[string][]byte{
			"0.0.0.0":         []byte{0, 0, 0, 0},
			"1.2.3.4":         []byte{1, 2, 3, 4},
			"192.168.34.56":   []byte{192, 168, 34, 56},
			"224.1.2.3":       []byte{224, 1, 2, 3},
			"255.255.255.255": []byte{255, 255, 255, 255},
		}
	)

	for _, ipv4AttrType := range getAttrsByCategory(ipv4Category) {
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
	goodData := [][]byte{
		[]byte{0, 0, 0, 0},
		[]byte{1, 2, 3, 4},
		[]byte{192, 168, 34, 56},
		[]byte{224, 1, 2, 3},
		[]byte{255, 255, 255, 255},
	}

	for _, ipv4AttrType := range getAttrsByCategory(ipv4Category) {
		for _, testData := range goodData {
			testAttr := ipv4Attribute{
				attrType: ipv4AttrType,
				attrData: testData,
			}
			err := testAttr.Validate()
			if err != nil {
				t.Fatalf(err.Error()+"\n"+"Supposed good data %s produced error for %s.",
					fmt.Sprintf("%v", []byte(testAttr.attrData)), AttrTypeString[ipv4AttrType])
			}
		}
	}
}

func TestIpv4Attribute_Validate_WithBadData(t *testing.T) {
	goodData := [][]byte{
		nil,
		[]byte{},
		[]byte{0},
		[]byte{0, 0},
		[]byte{0, 0, 0},
		[]byte{0, 0, 0, 0, 0},
	}

	for _, ipv4AttrType := range getAttrsByCategory(ipv4Category) {
		for _, testData := range goodData {
			testAttr := ipv4Attribute{
				attrType: ipv4AttrType,
				attrData: testData,
			}

			err := testAttr.Validate()
			if err == nil {
				t.Fatalf("Bad data %s in %s did not error.",
					fmt.Sprintf("%v", []byte(testAttr.attrData)), AttrTypeString[ipv4AttrType])
			}
		}
	}
}

func TestNewAttrBuilder_Ipv4(t *testing.T) {
	intData := uint32(3232238605)
	stringData := "192.168.12.13"
	byteData := []byte{192, 168, 12, 13}
	for _, ipv4AttrType := range getAttrsByCategory(ipv4Category) {
		expected := []byte{byte(ipv4AttrType), 6, 192, 168, 12, 13}
		byInt, err := NewAttrBuilder().SetType(ipv4AttrType).SetInt(intData).Build()
		if err != nil {
			t.Fatal(err)
		}
		byString, err := NewAttrBuilder().SetType(ipv4AttrType).SetString(stringData).Build()
		if err != nil {
			t.Fatal(err)
		}
		byByte, err := NewAttrBuilder().SetType(ipv4AttrType).SetBytes(byteData).Build()
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Compare(expected, MarshalAttribute(byInt)) != 0 {
			t.Fatal("Attributes don't match")
		}
		if bytes.Compare(byInt.Bytes(), byString.Bytes()) != 0 {
			t.Fatal("Attributes don't match")
		}
		if bytes.Compare(byString.Bytes(), byByte.Bytes()) != 0 {
			t.Fatal("Attributes don't match")
		}
		if bytes.Compare(byByte.Bytes(), byInt.Bytes()) != 0 {
			t.Fatal("Attributes don't match")
		}
	}
}

func TestIpv4Attribute_StringBadData(t *testing.T) {
	var (
		badStringTestData = []string{
			"hello",
			"0",
			"1",
			"4294967295",
			"4294967296",
			"4294967297",
			"1-1-1-1",
			"1:1:1:1",
			"1.256.3.4",
		}
	)

	for _, ipv4AttrType := range getAttrsByCategory(ipv4Category) {
		for _, s := range badStringTestData {
			_, err := NewAttrBuilder().SetType(ipv4AttrType).SetString(s).Build()
			if err == nil {
				t.Fatalf("setting IPv4 attribute data with `%s' did not error", s)
			}
		}
	}
}

