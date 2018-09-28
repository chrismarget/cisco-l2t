package attribute

import (
	"testing"
)

func TestStringString(t *testing.T) {
	attrTypesToTest := getAttrsByCategory(stringCategory)
	for _, v := range attrTypesToTest {
		data1 := attr{
			attrType: v,
			attrData: []byte{65, 0},
		}
		expected1 := stringStringPrefix + "A"
		result1, err := stringString(data1)
		if err != nil {
			t.Error(err)
		}
		if result1 != expected1 {
			t.Errorf("expected '%s', got '%s'", expected1, result1)
		}
	}

}
