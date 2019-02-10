package attribute

func getAttrsByCategory(category attrCategory) []attrType {
	var attrTypesToTest []attrType
	for k, v := range attrCategoryByType {
		if v == category {
			attrTypesToTest = append(attrTypesToTest, k)
		}
	}
	return attrTypesToTest
}

//func TestParseL2tAttr(t *testing.T) {
//	data1 := []byte{}
//	_, err := ParseL2tAttr(data1)
//	if err == nil {
//		t.Error("data1: empty l2tattr should have failed to parse")
//	}
//
//	var data2 []byte
//	for i := 0; i < 256; i++ {
//		data2 = append(data2, 0)
//	}
//	_, err = ParseL2tAttr(data2)
//	if err == nil {
//		t.Error("data2: oversize l2tattr should have failed to parse")
//	}
//
//	//data3 := []byte{1, 8, 0, 0, 0, 0, 0, 0}
//	//expected3 := Attr{
//	//	AttrType: attrType(1),
//	//	AttrData: []byte{0, 0, 0, 0, 0, 0},
//	//}
//	//	result3, err := ParseL2tAttr(data3)
//	//	if err != nil {
//	//		t.Error(err)
//	//	}
//	//	if !reflect.DeepEqual(result3, expected3) {
//	//		t.Error("test3 of ParseL2tAttr produced unexpected results")
//	//	}
//}
