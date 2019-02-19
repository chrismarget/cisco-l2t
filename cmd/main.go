package main

import (
	"fmt"
	"github.com/chrismarget/cisco-l2t/attribute"
	"github.com/chrismarget/cisco-l2t/communicate"
	"github.com/chrismarget/cisco-l2t/message"
	"log"
	"net"
	"os"
)

func main() {
	var a attribute.Attribute
	var err error

	builder := message.NewMsgBuilder()
	builder.SetType(message.RequestSrc)

	a, err = attribute.NewAttrBuilder().SetType(attribute.SrcMacType).SetString("0011.d9a5.2260").Build()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	builder.AddAttr(a)

	a, err = attribute.NewAttrBuilder().SetType(attribute.DstMacType).SetString("0000.0c9f.f00c").Build()
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

	msg := builder.Build()

	err = msg.Validate()
	if err != nil {
		fmt.Println(err)
		os.Exit(7)
	}

	response, respondent, err := communicate.Communicate(msg, &net.UDPAddr{IP: []byte{192, 168, 0, 254}})
	if err != nil {
		fmt.Println(err)
		os.Exit(8)
	}

	log.Println(respondent.IP.String())
	log.Println(message.MsgTypeToString[response.Type()])
	for _, a := range response.Attributes() {
		log.Println(attribute.AttrTypePrettyString[a.Type()], a.String())
	}
}
