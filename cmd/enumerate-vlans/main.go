package main

import (
	"flag"
	"fmt"
	"github.com/chrismarget/cisco-l2t/attribute"
	"github.com/chrismarget/cisco-l2t/message"
	"github.com/chrismarget/cisco-l2t/target"
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

	target, err := target.NewTarget().
		AddIp(net.ParseIP(flag.Arg(0))).
		Build()
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	var att attribute.Attribute
	var found []int


	// loop over all VLANs
	for i := 1; i <= 4094; i++ {

		builder := message.NewMsgBuilder()
		builder.SetType(message.RequestSrc)

		att, err = attribute.NewAttrBuilder().SetType(attribute.SrcMacType).SetString("ffff.ffff.ffff").Build()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		builder.SetAttr(att)

		att, err = attribute.NewAttrBuilder().SetType(attribute.DstMacType).SetString("ffff.ffff.ffff").Build()
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
		builder.SetAttr(att)

		att, err = attribute.NewAttrBuilder().SetType(attribute.VlanType).SetInt(uint32(i)).Build()
		if err != nil {
			fmt.Println(err)
			os.Exit(3)
		}
		builder.SetAttr(att)

		msg := builder.Build()

		err = msg.Validate()
		if err != nil {
			fmt.Println(err)
			os.Exit(7)
		}

		response, err := target.Send(msg)
		if err != nil {
			fmt.Println(err)
			os.Exit(8)
		}

		// Parse response. If we got the "good" error, add the VLAN to the list.
		for _, a := range response.Attributes() {
			if a.Type() == attribute.ReplyStatusType {
				if a.String() == "Source Mac address not found" {
					found = append(found, i)
				}
			}
		}
	}

	fmt.Printf("%d VLANs found:", len(found))
	var somefound bool
	// Pretty print results
	var a []int
	// iterate over found VLAN numbers
	for _, v := range found {
		somefound = true
		// Not the first one, right?
		if len(a) == 0 {
			// First VLAN. Initial slice population.
			a = append(a, v)
		} else {
			// Not the first VLAN, do sequential check
			if a[len(a)-1]+1 == v {
				// this VLAN is the next one in sequence
				a = append(a, v)
			} else {
				// there was a sequence gap, do some printing.
				if len(a) == 1 {
					// Just one number? Print it.
					fmt.Printf(" %d", a[0])
				} else {
					// More than one numbers? Print as a range.
					fmt.Printf(" %d-%d", a[0], a[len(a)-1])
				}
				a = []int{v}
			}
		}
	}
	if len(a) == 1 {
		// Just one number? Print it.
		fmt.Printf(" %d", a[0])
	} else {
		// More than one numbers? Print as a range.
		fmt.Printf(" %d-%d", a[0], a[len(a)-1])
	}
	if somefound {
		fmt.Printf(".\n")
	} else {
		fmt.Printf("<none>.\n")
	}

	//fmt.Printf("%d - %d", a[0], found[len(found)-1], "\n")

}
