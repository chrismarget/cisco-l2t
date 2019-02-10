package attribute

//func TestStringVlan(t *testing.T) {
//	attrTypesToTest := getAttrsByCategory(vlanCategory)
//	for _, v := range attrTypesToTest {
//		data1 := Attr{
//			AttrType: v,
//			AttrData: []byte{0, 10},
//		}
//		expected1 := "10"
//		result1, err := data1.String()
//		if err != nil {
//			t.Error(err)
//		}
//		if result1 != expected1 {
//			t.Errorf("expected '%s', got '%s'", expected1, result1)
//		}
//
//		data2 := Attr{
//			AttrType: v,
//			AttrData: []byte{15, 160},
//		}
//		expected2 := "4000"
//		result2, err := data2.String()
//		if err != nil {
//			t.Error(err)
//		}
//		if result2 != expected2 {
//			t.Errorf("expected '%s', got '%s'", expected2, result2)
//		}
//
//		data3 := Attr{
//			AttrType: v,
//			AttrData: []byte{100},
//		}
//		_, err = data3.String()
//		if err == nil {
//			t.Errorf("Undersize payload should have produced an error")
//		}
//
//		data4 := Attr{
//			AttrType: v,
//			AttrData: []byte{0, 0, 0},
//		}
//		_, err = data4.String()
//		if err == nil {
//			t.Errorf("Oversize payload should have produced an error")
//		}
//
//		data5 := Attr{
//			AttrType: v,
//			AttrData: []byte{0, 0},
//		}
//		_, err = data5.String()
//		if err == nil {
//			t.Errorf("Zero VLAN should have produced an error")
//		}
//
//		data6 := Attr{
//			AttrType: v,
//			AttrData: []byte{16, 0},
//		}
//		_, err = data6.String()
//		if err == nil {
//			t.Errorf("> 12-bit VLAN ID should have produced an error")
//		}
//	}
//}

//func TestNewVLANAttr(t *testing.T) {
//	attrTypesToTest := getAttrsByCategory(vlanCategory)
//	for _, testType := range attrTypesToTest {
//		var result Attr
//		var expected Attr
//		var err error
//
//		// VLAN 1
//		result, err = NewAttr(testType, attrPayload{intData: 1})
//		if err != nil {
//			t.Error(err)
//		}
//		expected = Attr{AttrType: testType, AttrData: []byte{0, 1}}
//		if !reflect.DeepEqual(result, expected) {
//			t.Error("Error: Structures don't match.")
//		}
//
//		// VLAN 4094
//		result, err = NewAttr(testType, attrPayload{intData: 4094})
//		if err != nil {
//			t.Error(err)
//		}
//		expected = Attr{AttrType: testType, AttrData: []byte{15, 254}}
//		if !reflect.DeepEqual(result, expected) {
//			t.Error("Error: Structures don't match.")
//		}
//
//		// VLAN 0
//		_, err = NewAttr(testType, attrPayload{intData: 0})
//		if err == nil {
//			t.Error("VLAN 0 should have produced an error.")
//		}
//
//		// VLAN 4095
//		_, err = NewAttr(testType, attrPayload{intData: 4095})
//		if err == nil {
//			t.Error("VLAN 4095 should have produced an error.")
//		}
//	}
//}
//
//func TestFoo(t *testing.T) {
//	o := vlan{attrType: 3, attrData: []byte{5, 5}}
//	err := o.validate()
//	if err != nil {
//		t.Fatal(err)
//	}
//	fmt.Println(o.string())
//
//}
