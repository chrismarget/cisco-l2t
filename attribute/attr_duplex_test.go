package attribute

import "testing"

func TestStringDuplex(t *testing.T) {
	attrTypesToTest := getAttrsByCategory(duplexCategory)
	for _, v := range attrTypesToTest {
		data1 := attr{
			attrType: v,
			attrData: []byte{0},
		}
		expected1 := duplexString[autoDuplex]
		result1, err := data1.String()
		if err != nil {
			t.Error(err)
		}
		if result1 != expected1 {
			t.Errorf("expected '%s', got '%s'", expected1, result1)
		}

		data2 := attr{
			attrType: v,
			attrData: []byte{1},
		}
		expected2 := duplexString[halfDuplex]
		result2, err := data2.String()
		if err != nil {
			t.Error(err)
		}
		if result2 != expected2 {
			t.Errorf("expected '%s', got '%s'", expected2, result2)
		}

		data3 := attr{
			attrType: v,
			attrData: []byte{2},
		}
		expected3 := duplexString[fullDuplex]
		result3, err := data3.String()
		if err != nil {
			t.Error(err)
		}
		if result3 != expected3 {
			t.Errorf("expected '%s', got '%s'", expected3, result3)
		}

		data4 := attr{
			attrType: v,
			attrData: []byte{3},
		}
		_, err = data4.String()
		if err == nil {
			t.Error("Bogus duplex value should have produced an error")
		}

		data5 := attr{
			attrType: v,
			attrData: []byte{0, 0},
		}
		_, err = data5.String()
		if err == nil {
			t.Error("Overlength duplex value should have produced an error")
		}

		data6 := attr{
			attrType: v,
			attrData: []byte{},
		}
		_, err = data6.String()
		if err == nil {
			t.Error("Empty duplex value should have produced an error")
		}
	}
}
