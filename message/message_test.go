package message

import (
	"bytes"
	"github.com/chrismarget/cisco-l2t/attribute"
	"testing"
)

func TestNewMsgBuilder_Minimal(t *testing.T) {
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

func TestNewMsgBuilder(t *testing.T) {
	const expectedLen = 109 // Header 5 plus attributes: 8+8+4+9+9+6+9+9+6+6+3+3+6+6+3+9

	var testType []attribute.AttrType
	var testString []string

	testType = append(testType, attribute.SrcMacType)
	testString = append(testString, "01:02:03:04:05:06")

	testType = append(testType, attribute.DstMacType)
	testString = append(testString, "02-03-04-05-06-07")

	testType = append(testType, attribute.VlanType)
	testString = append(testString, "257")

	testType = append(testType, attribute.DevNameType)
	testString = append(testString, "hello1")

	testType = append(testType, attribute.DevTypeType)
	testString = append(testString, "hello2")

	testType = append(testType, attribute.DevIPv4Type)
	testString = append(testString, "1.2.3.4")

	testType = append(testType, attribute.InPortNameType)
	testString = append(testString, "hello3")

	testType = append(testType, attribute.OutPortNameType)
	testString = append(testString, "hello4")

	testType = append(testType, attribute.InPortSpeedType)
	testString = append(testString, "10gbps")

	testType = append(testType, attribute.OutPortSpeedType)
	testString = append(testString, "100gb/s")

	testType = append(testType, attribute.InPortDuplexType)
	testString = append(testString, "auto")

	testType = append(testType, attribute.OutPortDuplexType)
	testString = append(testString, "half")

	testType = append(testType, attribute.NbrIPv4Type)
	testString = append(testString, "10.11.12.13")

	testType = append(testType, attribute.SrcIPv4Type)
	testString = append(testString, "20.21.22.23")

	testType = append(testType, attribute.ReplyStatusType)
	testString = append(testString, "Destination Mac address not found")

	testType = append(testType, attribute.NbrDevIDType)
	testString = append(testString, "hello5")

	var atts []attribute.Attribute
	for i, _ := range testType {
		a, err := attribute.NewAttrBuilder().SetType(testType[i]).SetString(testString[i]).Build()
		if err != nil {
			t.Fatal(err)
		}
		atts = append(atts, a)
	}

	builder := NewMsgBuilder()
	builder = builder.SetType(requestDst)
	for _, a := range atts {
		builder = builder.AddAttr(a)
	}

	msg, err := builder.Build()
	if err != nil {
		t.Fatal(err)
	}

	err = msg.Validate()
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewMsgBuilder_BadData(t *testing.T) {
	var testType []attribute.AttrType
	var testString []string

	testType = append(testType, attribute.SrcMacType)
	testString = append(testString, "01:02:03:04:05:06")

	testType = append(testType, attribute.DstMacType)
	testString = append(testString, "02-03-04-05-06-07")

	testType = append(testType, attribute.SrcMacType)
	testString = append(testString, "0506.0708.090a")

	var atts []attribute.Attribute
	for i, _ := range testType {
		a, err := attribute.NewAttrBuilder().SetType(testType[i]).SetString(testString[i]).Build()
		if err != nil {
			t.Fatal(err)
		}
		atts = append(atts, a)
	}

	builder := NewMsgBuilder()
	builder = builder.SetType(requestDst)
	for _, a := range atts {
		builder = builder.AddAttr(a)
	}

	msg, err := builder.Build()
	if err != nil {
		t.Fatal(err)
	}

	err = msg.Validate()
	if err == nil {
		t.Fatal("bad data should have provoked error")
	}
}

func TestMarshalMsg_Minimal(t *testing.T) {
	msg, err := NewMsgBuilder().Build()
	if err != nil {
		t.Fatal(err)
	}
	err = msg.Validate()
	if err != nil {
		t.Fatal(err)
	}
	expected := []byte{1, 1, 0, 5, 0}
	result := MarshalMsg(msg)
	if len(result) != len(expected) {
		t.Fatalf("expected 5 bytes")
	}

	if bytes.Compare(result, expected) != 0 {
		t.Fatalf("minimal marshaled message bad data")
	}
}
