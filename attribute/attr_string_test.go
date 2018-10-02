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
		expected1 := "A"
		result1, err := data1.String()
		if err != nil {
			t.Error(err)
		}
		if result1 != expected1 {
			t.Errorf("expected '%s', got '%s'", expected1, result1)
		}
	}

}
