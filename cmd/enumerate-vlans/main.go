package main

import (
	"flag"
	"fmt"
	"github.com/chrismarget/cisco-l2t/attribute"
	"github.com/chrismarget/cisco-l2t/communicate"
	"github.com/chrismarget/cisco-l2t/message"
	"log"
	"net"
	"os"
)

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		log.Println("You need to specify a target switch")
		os.Exit(1)
	}

	target := net.ParseIP(flag.Arg(0))
	if target == nil {
		log.Printf("bogus target: `%s'", flag.Arg(0))
		os.Exit(2)
	}

	var att attribute.Attribute
	var err error
	var found []int

	for i := 1; i <= 1; i++ {

		builder := message.NewMsgBuilder()
		builder.SetType(message.RequestSrc)

		att, err = attribute.NewAttrBuilder().SetType(attribute.SrcMacType).SetString("1000.1000.1000").Build()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		builder.AddAttr(att)

		att, err = attribute.NewAttrBuilder().SetType(attribute.DstMacType).SetString("2000.2000.2000").Build()
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
		builder.AddAttr(att)

		att, err = attribute.NewAttrBuilder().SetType(attribute.VlanType).SetInt(uint32(i)).Build()
		if err != nil {
			fmt.Println(err)
			os.Exit(3)
		}
		builder.AddAttr(att)

		msg := builder.Build()

		err = msg.Validate()
		if err != nil {
			fmt.Println(err)
			os.Exit(7)
		}

		log.Println(msg.Len())
		response, _, err := communicate.Communicate(msg, &net.UDPAddr{IP: target})
		log.Println(msg.Len())
		response, _, err = communicate.Communicate(msg, &net.UDPAddr{IP: target})
		log.Println(msg.Len())
		if err != nil {
			fmt.Println(err)
			os.Exit(8)
		}

		for _, a := range response.Attributes() {
			if a.Type() == attribute.ReplyStatusType {
				log.Println(i, a.String())
				if a.String() == "Source Mac address not found" {
					found = append(found, i)
					//log.Printf("%d", i)
				}
			}
		}
	}

	log.Println(found)

	var a []int
	for _, v := range found {
		if len(a) > 0 {
			if len(a) > 0 && a[len(a)-1]+1 == v {
				a = append(a, v)
			} else {
				if len(a) == 1 {
					fmt.Printf("%d ", a[0])
				} else {
					fmt.Printf("%d - %d ", a[0], a[len(a)-1])
				}
				a = append([]int{}, v)
			}
		} else {
			a = append(a, v)
		}
	}
	fmt.Printf("%d - %d", a[0], found[len(found)-1])

}
