package attribute

import "testing"

func TestStringIPv4(t *testing.T) {
	attrTypesToTest := getAttrsByCategory(ipv4Category)
	for _, v := range attrTypesToTest {
		data1 := attr{
			attrType: v,
			attrData: []byte{192, 168, 10, 11},
		}
		expected1 := "192.168.10.11"
		result1, err := stringIPv4(data1)
		if err != nil {
			t.Error(err)
		}
		if result1 != expected1 {
			t.Errorf("expected '%s', got '%s'", expected1, result1)
		}
	}
}
