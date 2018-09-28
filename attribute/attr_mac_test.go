package attribute

import "testing"

func TestStringMac(t *testing.T) {
	attrTypesToTest := getAttrsByCategory(macCategory)
	for _, v := range attrTypesToTest {
		data1 := attr{
			attrType: v,
			attrData: []byte{0, 0, 0, 0, 0, 0},
		}
		expected1 := macStringPrefix + "00:00:00:00:00:00"
		result1, err := stringMac(data1)
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
		expected2 := macStringPrefix + "ff:ff:ff:ff:ff:ff"
		result2, err := stringMac(data2)
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
		_, err = stringMac(data3)
		if err == nil {
			t.Error("Undersize MAC payload should have generated and error")
		}

		data4 := attr{
			attrType: v,
			attrData: []byte{0, 0, 0, 0, 0, 0, 0},
		}
		_, err = stringMac(data4)
		if err == nil {
			t.Error("Oversize MAC payload should have generated and error")
		}
	}
}
