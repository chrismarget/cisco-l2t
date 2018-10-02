package attribute

import "testing"

func TestStringMac(t *testing.T) {
	attrTypesToTest := getAttrsByCategory(macCategory)
	for _, v := range attrTypesToTest {
		data1 := attr{
			attrType: v,
			attrData: []byte{0, 0, 0, 0, 0, 0},
		}
		expected1 := "00:00:00:00:00:00"
		result1, err := data1.String()
		if err != nil {
			t.Error(err)
		}
		if result1 != expected1 {
			t.Errorf("expected '%s', got '%s'", expected1, result1)
		}

		data2 := attr{
			attrType: v,
			attrData: []byte{255, 255, 255, 255, 255, 255},
		}
		expected2 := "ff:ff:ff:ff:ff:ff"
		result2, err := data2.String()
		if err != nil {
			t.Error(err)
		}
		if result2 != expected2 {
			t.Errorf("expected '%s', got '%s'", expected2, result2)
		}

		data3 := attr{
			attrType: v,
			attrData: []byte{0, 0, 0, 0, 0},
		}
		_, err = data3.String()
		if err == nil {
			t.Error("Undersize MAC payload should have generated and error")
		}

		data4 := attr{
			attrType: v,
			attrData: []byte{0, 0, 0, 0, 0, 0, 0},
		}
		_, err = data4.String()
		if err == nil {
			t.Error("Oversize MAC payload should have generated and error")
		}
	}
}
