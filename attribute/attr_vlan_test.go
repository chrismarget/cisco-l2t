package attribute

import "testing"

func TestStringVlan(t *testing.T) {
	attrTypesToTest := getAttrsByCategory(vlanCategory)
	for _, v := range attrTypesToTest {
		data1 := attr{
			attrType: v,
			attrData: []byte{0, 10},
		}
		expected1 := "10"
		result1, err := stringVlan(data1)
		if err != nil {
			t.Error(err)
		}
		if result1 != expected1 {
			t.Errorf("expected '%s', got '%s'", expected1, result1)
		}

		data2 := attr{
			attrType: v,
			attrData: []byte{15, 160},
		}
		expected2 := "4000"
		result2, err := stringVlan(data2)
		if err != nil {
			t.Error(err)
		}
		if result2 != expected2 {
			t.Errorf("expected '%s', got '%s'", expected2, result2)
		}

		data3 := attr{
			attrType: v,
			attrData: []byte{100},
		}
		_, err = stringVlan(data3)
		if err == nil {
			t.Errorf("Undersize payload should have produced an error")
		}

		data4 := attr{
			attrType: v,
			attrData: []byte{0, 0, 0},
		}
		_, err = stringVlan(data4)
		if err == nil {
			t.Errorf("Oversize payload should have produced an error")
		}

		data5 := attr{
			attrType: v,
			attrData: []byte{0, 0},
		}
		_, err = stringVlan(data5)
		if err == nil {
			t.Errorf("Zero VLAN should have produced an error")
		}

		data6 := attr{
			attrType: v,
			attrData: []byte{16, 0},
		}
		_, err = stringVlan(data6)
		if err == nil {
			t.Errorf("> 12-bit VLAN ID should have produced an error")
		}
	}
}
