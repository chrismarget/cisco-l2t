package attribute

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
	"testing"
)

func TestVlanAttribute_String(t *testing.T) {
	vlanStringTestData := make(map[string][]byte, maxVLAN)

	// fill vlanStringTestData with values like:
	// "1025" -> []byte{04, 01}
	for i := minVLAN; i <= maxVLAN; i++ {
		b := make([]byte, 2)
		binary.BigEndian.PutUint16(b, uint16(i))
		vlanStringTestData[strconv.Itoa(i)] = b
	}

	_ = vlanStringTestData

	for _, vlanAttrType := range getAttrsByCategory(vlanCategory) {
		for expected, data := range vlanStringTestData {
			testAttr := vlanAttribute{
				attrType: vlanAttrType,
				attrData: data,
			}
			result := testAttr.String()
			if result != expected {
				t.Fatalf("expected %s, got %s", expected, result)
			}
		}
	}
}

func TestVlanAttribute_Validate_WithGoodData(t *testing.T) {
	var goodData [][]byte
	for i := minVLAN; i <= maxVLAN; i++ {
		b := make([]byte, 2)
		binary.BigEndian.PutUint16(b, uint16(i))
		goodData = append(goodData, b)
	}

	for _, vlanAttrType := range getAttrsByCategory(vlanCategory) {
		for _, testData := range goodData {
			testAttr := vlanAttribute{
				attrType: vlanAttrType,
				attrData: testData,
			}
			err := testAttr.Validate()
			if err != nil {
				t.Fatalf(err.Error()+"\n"+"Supposed good data %s produced error for %s.",
					fmt.Sprintf("%v", []byte(testAttr.attrData)), attrTypeString[vlanAttrType])
			}
		}
	}
}

func TestVlanAttribute_Validate_WithBadData(t *testing.T) {
	badData := [][]byte{
		nil,
		[]byte{},
		[]byte{0},
		[]byte{0, 0},
		[]byte{255, 255},
		[]byte{0, 0, 0},
	}

	for _, vlanAttrType := range getAttrsByCategory(vlanCategory) {
		for _, testData := range badData {
			testAttr := vlanAttribute{
				attrType: vlanAttrType,
				attrData: testData,
			}

			err := testAttr.Validate()
			if err == nil {
				t.Fatalf("Bad data %s in %s did not error.",
					fmt.Sprintf("%v", []byte(testAttr.attrData)), attrTypeString[vlanAttrType])
			}
		}
	}
}

func TestNewAttrBuilder_Vlan(t *testing.T) {
	for _, vlanAttrType := range getAttrsByCategory(vlanCategory) {
		var v uint16
		vBytes := make([]byte, 2)
		for v = 1; v <= 4094; v++ {
			binary.BigEndian.PutUint16(vBytes, v)
			expectedMarshal := []byte{byte(vlanAttrType), 4, vBytes[0], vBytes[1]}
			byInt, err := NewAttrBuilder().SetType(vlanAttrType).SetInt(uint32(v)).Build()
			if err != nil {
				t.Fatal(err)
			}
			byString, err := NewAttrBuilder().SetType(vlanAttrType).SetString(strconv.Itoa(int(v))).Build()
			if err != nil {
				t.Fatal(err)
			}
			byByte, err := NewAttrBuilder().SetType(vlanAttrType).SetBytes(vBytes).Build()
			if err != nil {
				t.Fatal(err)
			}

			if bytes.Compare(expectedMarshal, MarshalAttribute(byInt)) != 0 {
				t.Fatal("Attributes don't match")
			}
			if bytes.Compare(byInt.Bytes(), byByte.Bytes()) != 0 {
				t.Fatal("Attributes don't match")
			}
			if bytes.Compare(byByte.Bytes(), byString.Bytes()) != 0 {
				t.Fatal("Attributes don't match")
			}
			if bytes.Compare(byString.Bytes(), byInt.Bytes()) != 0 {
				t.Fatal("Attributes don't match")
			}
		}
	}
}
