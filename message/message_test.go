package message

import (
	"testing"
)

func TestNewMsgBuilder(t *testing.T) {
	msg, err := NewMsgBuilder().Build()
	if err != nil {
		t.Fatal(err)
	}
	if msg.Len() != 5 {
		t.Fatal("Default message should be 5 bytes")
	}
	if msg.Type() != requestDst {
		t.Fatalf("Default message type should be %s", msgTypeToString[requestDst])
	}
	if msg.AttrCount() != 0 {
		t.Fatal("Attribute count foa a default message should be zero")
	}
	if len(msg.Attributes()) != 0 {
		t.Fatal("Default message should have no attributes")
	}
	err = msg.Validate()
	if err != nil {
		t.Fatal(err)
	}
}

//func TestBytesToAttrSlice(t *testing.T) {
//	var result []attribute.Attr
//	var err error
//
//	// Test one attribute
//	result, err = bytesToAttrSlice([]byte{11, 3, 1})
//	if err != nil {
//		t.Error(err)
//	}
//	if len(result) != 1 {
//		t.Error("Unexpected result count")
//	}
//	if !reflect.DeepEqual(result[0], attribute.Attr{11, []byte{1}}) {
//		t.Error("Attributes don't match")
//	}
//
//	//// Test two attributes
//	//result, err = bytesToAttrSlice([]byte{1, 8, 0, 0, 0, 0, 0, 0, 2, 8, 2, 2, 2, 2, 2, 2})
//	//if err != nil {
//	//	t.Error(err)
//	//}
//	//if len(result) != 2 {
//	//	t.Error("Unexpected result count")
//	//}
//	//if !reflect.DeepEqual(result[0], attribute.Attr{1, []byte{0, 0, 0, 0, 0, 0}}) {
//	//	t.Error("Attributes don't match")
//	//}
//	//if !reflect.DeepEqual(result[1], attribute.Attr{2, []byte{2, 2, 2, 2, 2, 2}}) {
//	//	t.Error("Attributes don't match")
//	//}
//
//	// Test empty data
//	result, err = bytesToAttrSlice([]byte{})
//	if err != nil {
//		t.Error(err)
//	}
//	if len(result) != 0 {
//		t.Error("No input data should yield no attribute structures.")
//	}
//
//	// Test short data (one byte)
//	result, err = bytesToAttrSlice([]byte{1})
//	if err == nil {
//		t.Error("Single byte of input data should produce an error.")
//	}
//
//	// Test short data (two bytes)
//	result, err = bytesToAttrSlice([]byte{1, 2})
//	if err == nil {
//		t.Error("Two bytes of input data should produce an error.")
//	}
//
//	// Test short payload
//	result, err = bytesToAttrSlice([]byte{1, 8, 0, 0, 0, 0, 0})
//	if err == nil {
//		t.Error("Short payload should produce an error.")
//	}
//
//	// Test long payload
//	result, err = bytesToAttrSlice([]byte{1, 8, 0, 0, 0, 0, 0, 0, 0})
//	if err == nil {
//		t.Error("Long payload should produce an error.")
//	}
//}
