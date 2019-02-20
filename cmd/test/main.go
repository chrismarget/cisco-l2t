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

// enumerate vlans:
//   requestDst
//   both MACs set to ffff.ffff.ffff
//   iterate over vlans : 3 (no vlan) / 7 (vlan exists)
//

func main() {
	var a attribute.Attribute
	var err error

	builder := message.NewMsgBuilder()
	builder.SetType(message.RequestSrc)

	//a, err = attribute.NewAttrBuilder().SetType(attribute.SrcMacType).SetString("0030.18a0.1243").Build()
	a, err = attribute.NewAttrBuilder().SetType(attribute.SrcMacType).SetString("ffff.ffff.ffff").Build()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	builder.SetAttr(a)

	a, err = attribute.NewAttrBuilder().SetType(attribute.DstMacType).SetString("ffff.ffff.ffff").Build()
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	builder.SetAttr(a)

	a, err = attribute.NewAttrBuilder().SetType(attribute.VlanType).SetInt(43).Build()
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}
	builder.SetAttr(a)

	msg := builder.Build()

	err = msg.Validate()
	if err != nil {
		fmt.Println(err)
		os.Exit(7)
	}

	response, respondent, err := communicate.Communicate(msg, &net.UDPAddr{IP: []byte{192, 168, 96, 167}})
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
