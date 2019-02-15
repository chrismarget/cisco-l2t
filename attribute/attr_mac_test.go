package attribute

import (
	"bytes"
	"fmt"
	"testing"
)

func TestMacAttribute_String(t *testing.T) {
	var (
		macStringTestData = map[string][]byte{
			"00:00:00:00:00:00": []byte{0, 0, 0, 0, 0, 0},
			"01:02:03:04:05:06": []byte{1, 2, 3, 4, 5, 6},
			"ff:ff:ff:ff:ff:ff": []byte{255, 255, 255, 255, 255, 255},
		}
	)

	for _, macAttrType := range getAttrsByCategory(macCategory) {
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
	goodData := [][]byte{
		[]byte{0, 0, 0, 0, 0, 0},
		[]byte{1, 2, 3, 4, 5, 6},
		[]byte{255, 255, 255, 255, 255, 255},
	}

	for _, macAttrType := range getAttrsByCategory(macCategory) {
		for _, testData := range goodData {
			testAttr := macAttribute{
				attrType: macAttrType,
				attrData: testData,
			}
			err := testAttr.Validate()
			if err != nil {
				t.Fatalf(err.Error()+"\n"+"Supposed good data %s produced error for %s.",
					fmt.Sprintf("%v", []byte(testAttr.attrData)), AttrTypeString[macAttrType])
			}
		}
	}
}

func TestMacAttribute_Validate_WithBadData(t *testing.T) {
	badData := [][]byte{
		nil,
		[]byte{},
		[]byte{0},
		[]byte{0, 0},
		[]byte{0, 0, 0},
		[]byte{0, 0, 0, 0},
		[]byte{0, 0, 0, 0, 0},
		[]byte{0, 0, 0, 0, 0, 0, 0},
	}

	for _, macAttrType := range getAttrsByCategory(macCategory) {
		for _, testData := range badData {
			testAttr := macAttribute{
				attrType: macAttrType,
				attrData: testData,
			}

			err := testAttr.Validate()
			if err == nil {
				t.Fatalf("Bad data %s in %s did not error.",
					fmt.Sprintf("%v", []byte(testAttr.attrData)), AttrTypeString[macAttrType])
			}
		}
	}
}

func TestNewAttrBuilder_Mac(t *testing.T) {
	stringData1 := "00:01:02:FD:FE:FF"
	stringData2 := "00:01:02:fd:fe:ff"
	stringData3 := "00-01-02-FD-FE-FF"
	stringData4 := "00-01-02-fd-fe-ff"
	stringData5 := "0001.02FD.FEFF"
	stringData6 := "0001.02fd.feff"
	byteData := []byte{0, 1, 2, 253, 254, 255}
	for _, macAttrType := range getAttrsByCategory(macCategory) {
		expected := []byte{byte(macAttrType), 8, 0, 1, 2, 253, 254, 255}
		byString1, err := NewAttrBuilder().SetType(macAttrType).SetString(stringData1).Build()
		if err != nil {
			t.Fatal(err)
		}
		byString2, err := NewAttrBuilder().SetType(macAttrType).SetString(stringData2).Build()
		if err != nil {
			t.Fatal(err)
		}
		byString3, err := NewAttrBuilder().SetType(macAttrType).SetString(stringData3).Build()
		if err != nil {
			t.Fatal(err)
		}
		byString4, err := NewAttrBuilder().SetType(macAttrType).SetString(stringData4).Build()
		if err != nil {
			t.Fatal(err)
		}
		byString5, err := NewAttrBuilder().SetType(macAttrType).SetString(stringData5).Build()
		if err != nil {
			t.Fatal(err)
		}
		byString6, err := NewAttrBuilder().SetType(macAttrType).SetString(stringData6).Build()
		if err != nil {
			t.Fatal(err)
		}
		byByte, err := NewAttrBuilder().SetType(macAttrType).SetBytes(byteData).Build()
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Compare(expected, MarshalAttribute(byByte)) != 0 {
			t.Fatal("Attributes don't match")
		}
		if bytes.Compare(byByte.Bytes(), byString1.Bytes()) != 0 {
			t.Fatal("Attributes don't match")
		}
		if bytes.Compare(byString1.Bytes(), byString2.Bytes()) != 0 {
			t.Fatal("Attributes don't match")
		}
		if bytes.Compare(byString2.Bytes(), byString3.Bytes()) != 0 {
			t.Fatal("Attributes don't match")
		}
		if bytes.Compare(byString3.Bytes(), byString4.Bytes()) != 0 {
			t.Fatal("Attributes don't match")
		}
		if bytes.Compare(byString4.Bytes(), byString5.Bytes()) != 0 {
			t.Fatal("Attributes don't match")
		}
		if bytes.Compare(byString5.Bytes(), byString6.Bytes()) != 0 {
			t.Fatal("Attributes don't match")
		}
		if bytes.Compare(byString6.Bytes(), byByte.Bytes()) != 0 {
			t.Fatal("Attributes don't match")
		}
	}
}
