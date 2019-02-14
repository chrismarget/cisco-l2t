package attribute

import (
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

func TestUnMarshalAttribute(t *testing.T) {
	var testData = [][]byte{}
	var testType = []AttrType{}
	var testString = []string{}

	testData = append(testData, []byte{1, 8, 1, 2, 3, 4, 5, 6})
	testType = append(testType, SrcMacType)
	testString = append(testString, "01:02:03:04:05:06")

	testData = append(testData, []byte{2, 8, 2, 3, 4, 5, 6, 7})
	testType = append(testType, DstMacType)
	testString = append(testString, "02-03-04-05-06-07")

	testData = append(testData, []byte{3, 4, 1, 1})
	testType = append(testType, VlanType)
	testString = append(testString, "257")

	testData = append(testData, []byte{4, 8, 104, 101, 108, 108, 111, 0})
	testType = append(testType, DevNameType)
	testString = append(testString, "hello")

	for i, _ := range testData {
		result, err := UnmarshalV1Attribute(testData[i])
		if err != nil {
			t.Fatal(err)
		}

		expected, err := NewAttrBuilder().SetType(testType[i]).SetString(testString[i]).Build()
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(expected, result) {
			t.Fatal("structures don't match")
		}
	}
}

//for testType, _ := range(testData) {
//	result, err := NewAttrBuilder().SetType(SrcMacType).SetString().Build()
//	if err != nil {
//		t.Fatal(err)
//	}
//	expected, err := UnmarshalV1Attribute(testData[testType])
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
