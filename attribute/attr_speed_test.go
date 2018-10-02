package attribute

import (
	"encoding/binary"
	"testing"
)

func TestStringSpeed(t *testing.T) {
	attrTypesToTest := getAttrsByCategory(speedCategory)
	for _, v := range attrTypesToTest {
		speedVals := []uint32{0,1,2,3,4,5}
		expectedVals := []string{
			"Auto",
			"10Mb/s",
			"100Mb/s",
			"1Gb/s",
			"10Gb/s",
			"100Gb/s",
		}
		for k, _ := range(speedVals) {
			d := make([]byte, 4)
			binary.BigEndian.PutUint32(d, speedVals[k])
			data := attr{
				attrType: v,
				attrData: d,
			}
			result, err := stringSpeed(data)
			if err != nil {
				t.Error(err)
			}
			expected := expectedVals[k]
			if result != expected {
				t.Errorf("expected '%s', got '%s'", expected, result)
			}

		}
	}
}
