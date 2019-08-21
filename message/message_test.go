package message

import (
	"bytes"
	"github.com/chrismarget/cisco-l2t/attribute"
	"log"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func TestNewMsgBuilder_Minimal(t *testing.T) {
	msg := NewMsgBuilder().Build()
	if msg.Len() != 5 {
		t.Fatal("Default message should be 5 bytes")
	}
	if msg.Type() != RequestDst {
		t.Fatalf("Default message type should be %s", MsgTypeToString[RequestDst])
	}
	if msg.AttrCount() != 0 {
		t.Fatal("Attribute count foa a default message should be zero")
	}
	if len(msg.Attributes()) != 0 {
		t.Fatal("Default message should have no attributes")
	}
	err := msg.Validate()
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
	for i := range testType {
		a, err := attribute.NewAttrBuilder().SetType(testType[i]).SetString(testString[i]).Build()
		if err != nil {
			t.Fatal(err)
		}
		atts = append(atts, a)
	}

	builder := NewMsgBuilder()
	builder = builder.SetType(RequestDst)
	for _, a := range atts {
		builder = builder.SetAttr(a)
	}

	msg := builder.Build()

	err := msg.Validate()
	if err != nil {
		t.Fatal(err)
	}

	if msg.Len() != expectedLen {
		t.Fatalf("Expected %d, got %d", expectedLen, msg.Len())
	}
}

func TestLocationOfAttributeByType(t *testing.T) {
	testBytes := [][]byte{
		{0, 0, 0, 0, 0, 0},
		{1, 1, 1, 1, 1, 1},
		{2, 2},
		{33, 0},
		{34, 0},
	}

	testTypes := []attribute.AttrType{
		attribute.SrcMacType,
		attribute.DstMacType,
		attribute.VlanType,
		attribute.DevNameType,
		attribute.DevTypeType,
	}

	// build up a slice of attributes
	var testAttrs []attribute.Attribute
	for i := 0; i < len(testBytes); i++ {
		a, err := attribute.NewAttrBuilder().SetType(testTypes[i]).SetBytes(testBytes[i]).Build()
		if err != nil {
			t.Fatal(err)
		}
		testAttrs = append(testAttrs, a)
	}

	// throw those attributes in there again to be
	// sure we get only the first occurrence
	for i := 0; i < len(testBytes); i++ {
		a, err := attribute.NewAttrBuilder().SetType(testTypes[i]).SetBytes(testBytes[i]).Build()
		if err != nil {
			t.Fatal(err)
		}
		testAttrs = append(testAttrs, a)
	}

	expectedLocationByType := map[int]attribute.AttrType{
		0: attribute.SrcMacType,
		1: attribute.DstMacType,
		2: attribute.VlanType,
		3: attribute.DevNameType,
		4: attribute.DevTypeType,
	}

	for expected, attrType := range expectedLocationByType {
		result := attribute.LocationOfAttributeByType(testAttrs, attrType)
		if result != expected {
			t.Fatalf("expected %d, got %d", expected, result)
		}
	}
}

func TestOrderAttributes_ExactMatch(t *testing.T) {
	//attrs := make(map[attribute.AttrType]attribute.Attribute)
	var attrs []attribute.Attribute
	var a attribute.Attribute
	var err error
	a, err = attribute.NewAttrBuilder().SetType(16).SetString("foo").Build()
	if err != nil {
		t.Fatal(err)
	}
	attrs = append(attrs, a)
	a, err = attribute.NewAttrBuilder().SetType(3).SetInt(5).Build()
	if err != nil {
		t.Fatal(err)
	}
	attrs = append(attrs, a)
	a, err = attribute.NewAttrBuilder().SetType(2).SetBytes([]byte{2, 2, 2, 2, 2, 2}).Build()
	if err != nil {
		t.Fatal(err)
	}
	attrs = append(attrs, a)
	a, err = attribute.NewAttrBuilder().SetType(1).SetBytes([]byte{1, 1, 1, 1, 1, 1}).Build()
	if err != nil {
		t.Fatal(err)
	}
	attrs = append(attrs, a)
	a, err = attribute.NewAttrBuilder().SetType(14).SetBytes([]byte{1, 2, 3, 4}).Build()
	if err != nil {
		t.Fatal(err)
	}
	attrs = append(attrs, a)

	resultAttrs := orderAttributes(attrs, RequestDst)

	var resultTypes []attribute.AttrType
	for _, a := range resultAttrs {
		resultTypes = append(resultTypes, a.Type())
	}

	expected := []attribute.AttrType{2, 1, 3, 14, 16}

	if len(resultTypes) != len(expected) {
		t.Fatalf("results have unexpected length: got %d, expected %d", len(resultTypes), len(expected))
	}

	for i := range expected {
		if expected[i] != resultTypes[i] {
			t.Fatalf("position %d expected %d got %d", i, expected[i], resultTypes[i])
		}
	}
}

func TestOrderAttributes_WithExtras(t *testing.T) {
	//attrs := make(map[attribute.AttrType]attribute.Attribute)
	var attrs []attribute.Attribute
	var a attribute.Attribute
	var err error
	a, err = attribute.NewAttrBuilder().SetType(16).SetString("foo").Build()
	if err != nil {
		t.Fatal(err)
	}
	attrs = append(attrs, a)
	a, err = attribute.NewAttrBuilder().SetType(5).SetString("bar").Build()
	if err != nil {
		t.Fatal(err)
	}
	attrs = append(attrs, a)
	a, err = attribute.NewAttrBuilder().SetType(3).SetInt(5).Build()
	if err != nil {
		t.Fatal(err)
	}
	attrs = append(attrs, a)
	a, err = attribute.NewAttrBuilder().SetType(2).SetBytes([]byte{2, 2, 2, 2, 2, 2}).Build()
	if err != nil {
		t.Fatal(err)
	}
	attrs = append(attrs, a)
	a, err = attribute.NewAttrBuilder().SetType(1).SetBytes([]byte{1, 1, 1, 1, 1, 1}).Build()
	if err != nil {
		t.Fatal(err)
	}
	attrs = append(attrs, a)
	a, err = attribute.NewAttrBuilder().SetType(10).SetBytes([]byte{0, 0, 0, 1}).Build()
	if err != nil {
		t.Fatal(err)
	}
	attrs = append(attrs, a)
	a, err = attribute.NewAttrBuilder().SetType(14).SetBytes([]byte{1, 2, 3, 4}).Build()
	if err != nil {
		t.Fatal(err)
	}
	attrs = append(attrs, a)

	resultAttrs := orderAttributes(attrs, RequestDst)

	var resultTypes []attribute.AttrType
	for _, a := range resultAttrs {
		resultTypes = append(resultTypes, a.Type())
	}

	// type 0 here are placeholders for attribute types
	// with no order expectation
	expected := []attribute.AttrType{2, 1, 3, 14, 16, 0, 0}

	if len(resultTypes) != len(expected) {
		t.Fatalf("results have unexpected length: got %d, expected %d", len(resultTypes), len(expected))
	}

	for i := range expected {
		if expected[i] != resultTypes[i] {
			if expected[i] != 0 {
				t.Fatalf("position %d expected %d got %d", i, expected[i], resultTypes[i])
			}
		}
	}
}

func TestOrderAttributes_ShortlistAndExtras(t *testing.T) {
	//attrs := make(map[attribute.AttrType]attribute.Attribute)
	var attrs []attribute.Attribute
	var a attribute.Attribute
	var err error
	a, err = attribute.NewAttrBuilder().SetType(5).SetString("bar").Build()
	if err != nil {
		t.Fatal(err)
	}
	attrs = append(attrs, a)
	a, err = attribute.NewAttrBuilder().SetType(3).SetInt(5).Build()
	if err != nil {
		t.Fatal(err)
	}
	attrs = append(attrs, a)
	a, err = attribute.NewAttrBuilder().SetType(2).SetBytes([]byte{2, 2, 2, 2, 2, 2}).Build()
	if err != nil {
		t.Fatal(err)
	}
	attrs = append(attrs, a)
	a, err = attribute.NewAttrBuilder().SetType(10).SetBytes([]byte{0, 0, 0, 1}).Build()
	if err != nil {
		t.Fatal(err)
	}
	attrs = append(attrs, a)
	a, err = attribute.NewAttrBuilder().SetType(14).SetBytes([]byte{1, 2, 3, 4}).Build()
	if err != nil {
		t.Fatal(err)
	}
	attrs = append(attrs, a)

	resultAttrs := orderAttributes(attrs, RequestDst)

	var resultTypes []attribute.AttrType
	for _, a := range resultAttrs {
		resultTypes = append(resultTypes, a.Type())
	}

	expected := []attribute.AttrType{2, 3, 14, 5, 10}

	if len(resultTypes) != len(expected) {
		t.Fatalf("results have unexpected length: got %d, expected %d", len(resultAttrs), len(expected))
	}

	for i := range expected {
		if expected[i] != resultTypes[i] {
			t.Fatalf("position %d expected %d got %d", i, expected[i], resultAttrs[i])
		}
	}
}

func TestOrderAttributes_Shortlist(t *testing.T) {
	//attrs := make(map[attribute.AttrType]attribute.Attribute)
	var attrs []attribute.Attribute
	var a attribute.Attribute
	var err error

	a, err = attribute.NewAttrBuilder().SetType(16).SetString("foo").Build()
	if err != nil {
		t.Fatal(err)
	}
	attrs = append(attrs, a)
	a, err = attribute.NewAttrBuilder().SetType(3).SetInt(5).Build()
	if err != nil {
		t.Fatal(err)
	}
	attrs = append(attrs, a)
	a, err = attribute.NewAttrBuilder().SetType(1).SetBytes([]byte{1, 1, 1, 1, 1, 1}).Build()
	if err != nil {
		t.Fatal(err)
	}
	attrs = append(attrs, a)

	resultAttrs := orderAttributes(attrs, RequestDst)

	var resultTypes []attribute.AttrType
	for _, a := range resultAttrs {
		resultTypes = append(resultTypes, a.Type())
	}

	expected := []attribute.AttrType{1, 3, 16}

	if len(resultTypes) != len(expected) {
		t.Fatalf("results have unexpected length: got %d, expected %d", len(resultTypes), len(expected))
	}

	for i := range expected {
		if expected[i] != resultTypes[i] {
			t.Fatalf("position %d expected %d got %d", i, expected[i], resultTypes[i])
		}
	}
}

func TestAttrTypeLocationInSlice(t *testing.T) {
	testData := []attribute.AttrType{1, 3, 5, 7, 9, 11, 13, 2, 4, 6, 8, 10, 12}
	expected := []int{-1, 1, 5, 11, -1}
	var result []int
	result = append(result, attrTypeLocationInSlice(testData, 16))
	result = append(result, attrTypeLocationInSlice(testData, 3))
	result = append(result, attrTypeLocationInSlice(testData, 11))
	result = append(result, attrTypeLocationInSlice(testData, 10))
	result = append(result, attrTypeLocationInSlice(testData, 15))

	if len(result) != len(expected) {
		t.Fatalf("expected %d results, got %d", len(expected), len(result))
	}

	for i := range expected {
		if expected[i] != result[i] {
			t.Fatalf("result %d expected %d got %d", i, expected[i], result[i])
		}
	}

}

func TestMarshalMsg_Minimal(t *testing.T) {
	msg := NewMsgBuilder().Build()

	err := msg.Validate()
	if err != nil {
		t.Fatal(err)
	}

	expected := []byte{1, 1, 0, 5, 0}
	result := msg.Marshal(nil)
	if len(result) != len(expected) {
		t.Fatalf("expected 5 bytes")
	}

	if bytes.Compare(result, expected) != 0 {
		t.Fatalf("minimal marshaled message bad data")
	}
}

func TestMarshalMsg_ReqDstStandard(t *testing.T) {
	var a attribute.Attribute
	var err error

	builder := NewMsgBuilder()
	builder.SetType(RequestDst)
	builder.SetVer(Version1)

	// Attribute should be {2, 8, 1, 2, 3, 4, 5, 6}
	a, err = attribute.NewAttrBuilder().SetType(attribute.DstMacType).SetString("0102.0304.0506").Build()
	if err != nil {
		t.Fatal(err)
	}
	builder.SetAttr(a)

	// Attribute should be {3, 4, 12, 34}
	a, err = attribute.NewAttrBuilder().SetType(attribute.VlanType).SetInt(3106).Build()
	if err != nil {
		t.Fatal(err)
	}
	builder.SetAttr(a)

	// Attribute should be {16, 6, 102, 111, 111, 0}
	a, err = attribute.NewAttrBuilder().SetType(attribute.NbrDevIDType).SetString("foo").Build()
	if err != nil {
		t.Fatal(err)
	}
	builder.SetAttr(a)

	// Attribute should be {14, 6, 1, 2, 3, 4}
	a, err = attribute.NewAttrBuilder().SetType(attribute.SrcIPv4Type).SetString("1.2.3.4").Build()
	if err != nil {
		t.Fatal(err)
	}
	builder.SetAttr(a)

	// Attribute should be {1, 8, 255, 254, 253, 5, 6, 7}
	a, err = attribute.NewAttrBuilder().SetType(attribute.SrcMacType).SetString("ff-fe-fd-05-06-07").Build()
	if err != nil {
		t.Fatal(err)
	}
	builder.SetAttr(a)

	msg := builder.Build()
	if err != nil {
		t.Fatal(err)
	}

	err = msg.Validate()
	if err != nil {
		t.Fatal(err)
	}

	// RequestDst header should be {2, 1, 0, 37, 5}
	// DstMacType   attribute should be {2, 8, 1, 2, 3, 4, 5, 6}
	// SrcMacType   attribute should be {1, 8, 255, 254, 253, 5, 6, 7}
	// VlanType     attribute should be {3, 4, 12, 34}
	// SrcIPv4Type  attribute should be {14, 6, 1, 2, 3, 4}
	// NbrDevIDType attribute should be {16, 6, 102, 111, 111, 0}

	expected := []byte{
		1, 1, 0, 37, 5,
		2, 8, 1, 2, 3, 4, 5, 6,
		1, 8, 255, 254, 253, 5, 6, 7,
		3, 4, 12, 34,
		14, 6, 1, 2, 3, 4,
		16, 6, 102, 111, 111, 0,
	}
	result := msg.Marshal(nil)
	if len(result) != len(expected) {
		t.Fatalf("got %d bytes, expected %d", len(result), len(expected))
	}

	if bytes.Compare(result, expected) != 0 {
		t.Fatalf("RequestDst standard message bad data")
	}
}

func TestMarshalMsg_ReqDstOversize(t *testing.T) {
	var a attribute.Attribute
	var err error

	builder := NewMsgBuilder()
	builder.SetType(RequestDst)
	builder.SetVer(Version1)

	// Attribute should be {2, 8, 1, 2, 3, 4, 5, 6}
	a, err = attribute.NewAttrBuilder().SetType(attribute.DstMacType).SetString("0102.0304.0506").Build()
	if err != nil {
		t.Fatal(err)
	}
	builder.SetAttr(a)

	// Attribute should be {3, 4, 12, 34}
	a, err = attribute.NewAttrBuilder().SetType(attribute.VlanType).SetInt(3106).Build()
	if err != nil {
		t.Fatal(err)
	}
	builder.SetAttr(a)

	// Attribute should be {16, 6, 102, 111, 111, 0}
	a, err = attribute.NewAttrBuilder().SetType(attribute.NbrDevIDType).SetString("foo").Build()
	if err != nil {
		t.Fatal(err)
	}
	builder.SetAttr(a)

	// Superflous Attribute should be {11, 3, 0}
	a, err = attribute.NewAttrBuilder().SetType(attribute.InPortDuplexType).SetInt(0).Build()
	if err != nil {
		t.Fatal(err)
	}
	builder.SetAttr(a)

	// Attribute should be {14, 6, 1, 2, 3, 4}
	a, err = attribute.NewAttrBuilder().SetType(attribute.SrcIPv4Type).SetString("1.2.3.4").Build()
	if err != nil {
		t.Fatal(err)
	}
	builder.SetAttr(a)

	// Attribute should be {1, 8, 255, 254, 253, 5, 6, 7}
	a, err = attribute.NewAttrBuilder().SetType(attribute.SrcMacType).SetString("ff-fe-fd-05-06-07").Build()
	if err != nil {
		t.Fatal(err)
	}
	builder.SetAttr(a)

	msg := builder.Build()
	if err != nil {
		t.Fatal(err)
	}

	err = msg.Validate()
	if err != nil {
		t.Fatal(err)
	}

	// RequestDst header should be {2, 1, 0, 40, 6}
	// DstMacType       attribute should be {2, 8, 1, 2, 3, 4, 5, 6}
	// SrcMacType       attribute should be {1, 8, 255, 254, 253, 5, 6, 7}
	// VlanType         attribute should be {3, 4, 12, 34}
	// SrcIPv4Type      attribute should be {14, 6, 1, 2, 3, 4}
	// NbrDevIDType     attribute should be {16, 6, 102, 111, 111, 0}
	// Superfluous
	// InPortDuplexType attribute should be {11, 3, 0}

	expected := []byte{
		1, 1, 0, 40, 6,
		2, 8, 1, 2, 3, 4, 5, 6,
		1, 8, 255, 254, 253, 5, 6, 7,
		3, 4, 12, 34,
		14, 6, 1, 2, 3, 4,
		16, 6, 102, 111, 111, 0,
		11, 3, 0,
	}
	result := msg.Marshal(nil)
	if len(result) != len(expected) {
		t.Fatalf("got %d bytes, expected %d", len(expected), len(result))
	}

	if bytes.Compare(result, expected) != 0 {
		t.Fatalf("RequestDst oversize message bad data")
	}
}

func TestMarshalMsg_ReqDstUndersize(t *testing.T) {
	var a attribute.Attribute
	var err error

	builder := NewMsgBuilder()
	builder.SetType(RequestDst)
	builder.SetVer(Version1)

	// Attribute should be {2, 8, 1, 2, 3, 4, 5, 6}
	a, err = attribute.NewAttrBuilder().SetType(attribute.DstMacType).SetString("0102.0304.0506").Build()
	if err != nil {
		t.Fatal(err)
	}
	builder.SetAttr(a)

	// Attribute should be {3, 4, 12, 34}
	a, err = attribute.NewAttrBuilder().SetType(attribute.VlanType).SetInt(3106).Build()
	if err != nil {
		t.Fatal(err)
	}
	builder.SetAttr(a)

	// Attribute should be {16, 6, 102, 111, 111, 0}
	a, err = attribute.NewAttrBuilder().SetType(attribute.NbrDevIDType).SetString("foo").Build()
	if err != nil {
		t.Fatal(err)
	}
	builder.SetAttr(a)

	// Attribute should be {1, 8, 255, 254, 253, 5, 6, 7}
	a, err = attribute.NewAttrBuilder().SetType(attribute.SrcMacType).SetString("ff-fe-fd-05-06-07").Build()
	if err != nil {
		t.Fatal(err)
	}
	builder.SetAttr(a)

	msg := builder.Build()
	if err != nil {
		t.Fatal(err)
	}

	err = msg.Validate()
	if err != nil {
		t.Fatal(err)
	}

	// RequestDst header should be {2, 1, 0, 31, 4}
	// DstMacType   attribute should be {2, 8, 1, 2, 3, 4, 5, 6}
	// SrcMacType   attribute should be {1, 8, 255, 254, 253, 5, 6, 7}
	// VlanType     attribute should be {3, 4, 12, 34}
	// NbrDevIDType attribute should be {16, 6, 102, 111, 111, 0}

	expected := []byte{
		1, 1, 0, 31, 4,
		2, 8, 1, 2, 3, 4, 5, 6,
		1, 8, 255, 254, 253, 5, 6, 7,
		3, 4, 12, 34,
		16, 6, 102, 111, 111, 0,
	}
	result := msg.Marshal(nil)
	if len(result) != len(expected) {
		t.Fatalf("got %d bytes, expected %d", len(expected), len(result))
	}

	if bytes.Compare(result, expected) != 0 {
		t.Fatalf("RequestDst undersize message bad data")
	}
}

func TestUnmarshalMessage(t *testing.T) {
	data := []byte{
		3, 1, 0, 86, 12,
		4, 10, 67, 97, 116, 50, 57, 54, 48, 0,
		5, 19, 87, 83, 45, 67, 50, 57, 54, 48, 71, 45, 50, 52, 84, 67, 45, 76, 0,
		6, 6, 192, 168, 1, 254,
		15, 3, 1,
		13, 6, 0, 0, 0, 0,
		16, 3, 0,
		7, 8, 71, 105, 48, 47, 52, 0,
		11, 3, 0,
		9, 6, 0, 0, 0, 0,
		8, 8, 71, 105, 48, 47, 52, 0,
		12, 3, 0,
		10, 6, 0, 0, 0, 0,
	}

	msg, err := UnmarshalMessage(data)
	if err != nil {
		t.Fatal(err)
	}

	_ = msg
	// TODO inspect the message
}

func TestString(t *testing.T) {
	data := []byte{
		3, 1, 0, 86, 12,
		4, 10, 67, 97, 116, 50, 57, 54, 48, 0,
		5, 19, 87, 83, 45, 67, 50, 57, 54, 48, 71, 45, 50, 52, 84, 67, 45, 76, 0,
		6, 6, 192, 168, 1, 254,
		15, 3, 1,
		13, 6, 0, 0, 0, 0,
		16, 3, 0,
		7, 8, 71, 105, 48, 47, 52, 0,
		11, 3, 0,
		9, 6, 0, 0, 0, 0,
		8, 8, 71, 105, 48, 47, 52, 0,
		12, 3, 0,
		10, 6, 0, 0, 0, 0,
	}

	msg, err := UnmarshalMessage(data)
	if err != nil {
		t.Fatal(err)
	}

	expected := "L2T_REPLY_DST with 12 attributes (86 bytes)"
	result := msg.String()
	if expected != result {
		t.Fatalf("expected '%s', got '%s'", expected, result)
	}
}

func TestSetAttr(t *testing.T) {
	msgData := []byte{
		2, 1, 0, 25, 3, // header	t2, v1, l25, ac3
		2, 8, 255, 255, 255, 255, 255, 255, // src_mac ff:ff:ff:ff:ff:ff
		1, 8, 255, 255, 255, 255, 255, 255, // dst_mac ff:ff:ff:ff:ff:ff
		3, 4, 0, 1, // vlan 1
	}

	msg, err := UnmarshalMessage(msgData)
	if err != nil {
		t.Fatal(err)
	}

	if msg.AttrCount() != 3 {
		t.Fatalf("expected attrcout 3, got %d", msg.AttrCount())
	}

	if msg.NeedsSrcIp() != true {
		t.Fatalf("message should have needed src ip")
	}

	srcIpAttr, err := attribute.NewAttrBuilder().
		SetType(attribute.SrcIPv4Type).
		SetString("1.1.1.1").
		Build()
	if err != nil {
		t.Fatal(err)
	}

	msg.SetAttr(srcIpAttr)

	if msg.AttrCount() != 4 {
		t.Fatalf("expected attrcout 4, got %d", msg.AttrCount())
	}

	_ = msg
}

func TestSetAttrAgain(t *testing.T) {
	msgData := []byte{
		2, 1, 0, 25, 3, // header	t2, v1, l25, ac3
		2, 8, 255, 255, 255, 255, 255, 255, // src_mac ff:ff:ff:ff:ff:ff
		1, 8, 255, 255, 255, 255, 255, 255, // dst_mac ff:ff:ff:ff:ff:ff
		3, 4, 0, 1, // vlan 1
	}

	msg, err := UnmarshalMessage(msgData)
	if err != nil {
		t.Fatal(err)
	}

	if msg.AttrCount() != 3 {
		t.Fatalf("expected attrcout 3, got %d", msg.AttrCount())
	}

	if msg.NeedsSrcIp() != true {
		t.Fatalf("message should have needed src ip")
	}

	srcIpAttr, err := attribute.NewAttrBuilder().
		SetType(attribute.SrcIPv4Type).
		SetString("1.1.1.1").
		Build()
	if err != nil {
		t.Fatal(err)
	}
	msg.SetAttr(srcIpAttr)

	rand.Seed(time.Now().UTC().UnixNano())

	v1 := (rand.Int() % 4093) + 1
	vlan1Attr, err := attribute.NewAttrBuilder().
		SetType(attribute.VlanType).
		SetInt(uint32(v1)).
		Build()
	if err != nil {
		log.Fatal(err)
	}

	v2 := (rand.Int() % 4093) + 1
	vlan2Attr, err := attribute.NewAttrBuilder().
		SetType(attribute.VlanType).
		SetInt(uint32(v2)).
		Build()
	if err != nil {
		log.Fatal(err)
	}

	msg.SetAttr(vlan1Attr)
	for _, a := range msg.Attributes() {
		if a.Type() == attribute.VlanType {
			if a.String() != strconv.Itoa(v1) {
				t.Fatalf("v1 got vlan %s, expected %d", a.String(), v1)
			}
		}
	}
	msg.SetAttr(vlan2Attr)
	for _, a := range msg.Attributes() {
		if a.Type() == attribute.VlanType {
			if a.String() != strconv.Itoa(v2) {
				t.Fatalf("v2 got vlan %s, expected %d", a.String(), v2)
			}
		}
	}

	if msg.AttrCount() != 4 {
		t.Fatalf("expected attrcout 4, got %d", msg.AttrCount())
	}
}
