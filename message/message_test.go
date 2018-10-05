package message

import (
	"github.com/chrismarget/cisco-l2t/attribute"
	"reflect"
	"testing"
)

func TestBytesToAttrSlice(t *testing.T) {
	var result []attribute.Attr
	var err error

	// Test one attribute
	result, err = bytesToAttrSlice([]byte{11, 3, 1})
	if err != nil {
		t.Error(err)
	}
	if len(result) != 1 {
		t.Error("Unexpected result count")
	}
	if !reflect.DeepEqual(result[0], attribute.Attr{11, []byte{1}}) {
		t.Error("Attributes don't match")
	}

	// Test two attributes
	result, err = bytesToAttrSlice([]byte{1, 8, 0, 0, 0, 0, 0, 0, 2, 8, 2, 2, 2, 2, 2, 2})
	if err != nil {
		t.Error(err)
	}
	if len(result) != 2 {
		t.Error("Unexpected result count")
	}
	if !reflect.DeepEqual(result[0], attribute.Attr{1, []byte{0, 0, 0, 0, 0, 0}}) {
		t.Error("Attributes don't match")
	}
	if !reflect.DeepEqual(result[1], attribute.Attr{2, []byte{2, 2, 2, 2, 2, 2}}) {
		t.Error("Attributes don't match")
	}

	// Test empty data
	result, err = bytesToAttrSlice([]byte{})
	if err != nil {
		t.Error(err)
	}
	if len(result) != 0 {
		t.Error("No input data should yield no attribute structures.")
	}

	// Test short data (one byte)
	result, err = bytesToAttrSlice([]byte{1})
	if err == nil {
		t.Error("Single byte of input data should produce an error.")
	}

	// Test short data (two bytes)
	result, err = bytesToAttrSlice([]byte{1, 2})
	if err == nil {
		t.Error("Two bytes of input data should produce an error.")
	}

	// Test short payload
	result, err = bytesToAttrSlice([]byte{1, 8, 0, 0, 0, 0, 0})
	if err == nil {
		t.Error("Short payload should produce an error.")
	}

	// Test long payload
	result, err = bytesToAttrSlice([]byte{1, 8, 0, 0, 0, 0, 0, 0, 0})
	if err == nil {
		t.Error("Long payload should produce an error.")
	}
}
