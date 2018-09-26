package l2t

import (
	"testing"
)

func TestMakePortDuplex(t *testing.T) {
	MakePortDuplex(0)
}

func TestStringMac(t *testing.T) {
	data1 := []byte{0,0,0,0,0,0}
	expected1 := "00:00:00:00:00:00"
	result1, err := stringMac(data1)
	if err != nil {
		t.Error(err)
	}
	if result1 != expected1 {
		t.Error("expected '%s', got '%s'", expected1, result1 )
	}
}