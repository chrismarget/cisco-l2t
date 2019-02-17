package main

import (
	"fmt"
	"github.com/chrismarget/cisco-l2t/attribute"
	"github.com/chrismarget/cisco-l2t/message"
	"log"
	"net"
	"os"
)

func main() {
	var a attribute.Attribute
	var err error

	builder := message.NewMsgBuilder()

	a, err = attribute.NewAttrBuilder().SetType(attribute.SrcMacType).SetString("7073.cb8a.f62a").Build()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	builder.AddAttr(a)

	a, err = attribute.NewAttrBuilder().SetType(attribute.DstMacType).SetString("5082.d5c6.e2d4").Build()
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	builder.AddAttr(a)

	a, err = attribute.NewAttrBuilder().SetType(attribute.VlanType).SetInt(12).Build()
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}
	builder.AddAttr(a)

	a, err = attribute.NewAttrBuilder().SetType(attribute.SrcIPv4Type).SetString("192.168.2.214").Build()
	if err != nil {
		fmt.Println(err)
		os.Exit(4)
	}
	builder.AddAttr(a)

	msg, err := builder.Build()
	if err != nil {
		fmt.Println(err)
		os.Exit(6)
	}

	err = msg.Validate()
	if err != nil {
		fmt.Println(err)
		os.Exit(7)
	}

	response, respondent, err := msg.Communicate(&net.UDPAddr{IP :[]byte{192,168,0,254}})
	if err != nil {
		fmt.Println(err)
		os.Exit(8)
	}

	log.Println(respondent.IP)
	log.Println(message.MsgTypeToString[response.Type()])
	for _, a := range response.Attributes() {
		log.Println(attribute.AttrTypePrettyString[a.Type()], a.String())
	}
}
