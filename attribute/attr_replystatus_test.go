package attribute

import (
	"strconv"
	"testing"
)

func TestStringStatus(t *testing.T) {
	for i := 0; i <= 255; i++ {
		data := attr{
			attrType: replyStatusType,
			attrData: []byte{byte(i)},
		}

		var expected string
		switch i {
		case 1:
			expected = "1 Success"
		case 9:
			expected = "9 No CDP Neighbor"
		default:
			expected = strconv.Itoa(i) + " Unknown"
		}

		result, err := data.String()
			if err != nil {
				t.Error(err)
			}
		if result != expected {
			t.Errorf("expected '%s', got '%s'", expected, result)
		}
	}
}
