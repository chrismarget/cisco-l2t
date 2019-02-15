package attribute

import (
	"math"
	"reflect"
	"testing"
)

func getAttrsByCategory(category attrCategory) []AttrType {
	var attrTypesToTest []AttrType
	for k, v := range attrCategoryByType {
		if v == category {
			attrTypesToTest = append(attrTypesToTest, k)
		}
	}
	return attrTypesToTest
}

func TestUnMarshalAttribute_BadData(t *testing.T) {
	var testData [][]byte
	testData = append(testData, []byte{})
	for i := 0; i <= math.MaxUint8; i++ {
		testData[0] = append(testData[0], 1)			// oversize (fill 1st instance to 256 bytes)
	}
	testData = append(testData, []byte{})						// undersize
	testData = append(testData, []byte{1})						// undersize
	testData = append(testData, []byte{1, 2})					// undersize
	testData = append(testData, []byte{1, 8, 0, 0, 0, 0, 0})	// wrong length

	for t := 0; t <= math.MaxUint8; t++ {
		if _, ok := attrCategoryByType[AttrType(t)]; !ok {
			testData = append(testData, []byte{byte(t), 3, 1})	// fill testData with bogus types
		}
	}

	for d, _ := range testData {
		_, err := UnmarshalAttribute(testData[d])
		if err == nil {
			t.Fatalf("bad data should have produced an error")
		}
	}
}

func TestUnMarshalAttribute(t *testing.T) {
	var testData [][]byte
	var testType []AttrType
	var testString []string

	testData = append(testData, []byte{1, 8, 1, 2, 3, 4, 5, 6})
	testType = append(testType, SrcMacType)
	testString = append(testString, "01:02:03:04:05:06")

	testData = append(testData, []byte{2, 8, 2, 3, 4, 5, 6, 7})
	testType = append(testType, DstMacType)
	testString = append(testString, "02-03-04-05-06-07")

	testData = append(testData, []byte{3, 4, 1, 1})
	testType = append(testType, VlanType)
	testString = append(testString, "257")

	testData = append(testData, []byte{4, 9, 104, 101, 108, 108, 111, 49, 0})
	testType = append(testType, DevNameType)
	testString = append(testString, "hello1")

	testData = append(testData, []byte{5, 9, 104, 101, 108, 108, 111, 50, 0})
	testType = append(testType, DevTypeType)
	testString = append(testString, "hello2")

	testData = append(testData, []byte{6, 6, 1, 2, 3, 4})
	testType = append(testType, DevIPv4Type)
	testString = append(testString, "1.2.3.4")

	testData = append(testData, []byte{7, 9, 104, 101, 108, 108, 111, 51, 0})
	testType = append(testType, InPortNameType)
	testString = append(testString, "hello3")

	testData = append(testData, []byte{8, 9, 104, 101, 108, 108, 111, 52, 0})
	testType = append(testType, OutPortNameType)
	testString = append(testString, "hello4")

	testData = append(testData, []byte{9, 6, 0, 0, 0, 4})
	testType = append(testType, InPortSpeedType)
	testString = append(testString, "10gbps")

	testData = append(testData, []byte{10, 6, 0, 0, 0, 5})
	testType = append(testType, OutPortSpeedType)
	testString = append(testString, "100gb/s")

	testData = append(testData, []byte{11, 3, 0})
	testType = append(testType, InPortDuplexType)
	testString = append(testString, "auto")

	testData = append(testData, []byte{12, 3, 1})
	testType = append(testType, OutPortDuplexType)
	testString = append(testString, "half")

	testData = append(testData, []byte{13, 6, 10, 11, 12, 13})
	testType = append(testType, NbrIPv4Type)
	testString = append(testString, "10.11.12.13")

	testData = append(testData, []byte{14, 6, 20, 21, 22, 23})
	testType = append(testType, SrcIPv4Type)
	testString = append(testString, "20.21.22.23")

	testData = append(testData, []byte{15, 3, 8})
	testType = append(testType, ReplyStatusType)
	testString = append(testString, "Destination Mac address not found")

	testData = append(testData, []byte{16, 9, 104, 101, 108, 108, 111, 53, 0})
	testType = append(testType, NbrDevIDType)
	testString = append(testString, "hello5")

	for i, _ := range testData {
		result, err := UnmarshalAttribute(testData[i])
		if err != nil {
			t.Fatal(err)
		}

		expected, err := NewAttrBuilder().SetType(testType[i]).SetString(testString[i]).Build()
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(expected, result) {
			t.Fatalf("structures don't match")
		}
	}
}

//for testType, _ := range(testData) {
//	result, err := NewAttrBuilder().SetType(SrcMacType).SetString().Build()
//	if err != nil {
//		t.Fatal(err)
//	}
//	expected, err := UnmarshalAttribute(testData[testType])
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	if ! reflect.DeepEqual(result, expected) {
//		log.Println(expected)
//		log.Println(result)
//		//log.Println(expected.Bytes())
//		//log.Println(result.Bytes())
//		t.Fatalf("Unmarshaled attribute does not match expected data")
//	}
//
//}

//var testData = map[Attribute][]byte{}

//SrcMacType
//DstMacType
//VlanType
//DevNameType
//DevTypeType
//DevIPv4Type
//InPortNameType
//OutPortNameType
//InPortSpeedType
//OutPortSpeedType
//InPortDuplexType
//OutPortDuplexType
//NbrIPv4Type
//SrcIPv4Type
//ReplyStatusType
//NbrDevIDType
