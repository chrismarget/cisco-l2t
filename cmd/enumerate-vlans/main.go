package main

import (
	"flag"
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"github.com/chrismarget/cisco-l2t/attribute"
	"github.com/chrismarget/cisco-l2t/message"
	"github.com/chrismarget/cisco-l2t/target"
	"log"
	"net"
	"os"
)

const (
	vlanMin = 1
	vlanMax = 4094
)

type vlan int

func enumerate_vlans(t target.Target) ([]vlan, error) {
	var att attribute.Attribute
	var found []vlan
	var err error

	bar := pb.StartNew(vlanMax)

	// loop over all VLANs
	for i := vlanMin; i <= vlanMax; i++ {
		bar.Increment()

		builder := message.NewMsgBuilder()
		builder.SetType(message.RequestSrc)

		att, err = attribute.NewAttrBuilder().SetType(attribute.SrcMacType).SetString("ffff.ffff.ffff").Build()
		if err != nil {
			return nil, err
		}
		builder.SetAttr(att)

		att, err = attribute.NewAttrBuilder().SetType(attribute.DstMacType).SetString("ffff.ffff.ffff").Build()
		if err != nil {
			return nil, err
		}
		builder.SetAttr(att)

		att, err = attribute.NewAttrBuilder().SetType(attribute.VlanType).SetInt(uint32(i)).Build()
		if err != nil {
			return nil, err
		}
		builder.SetAttr(att)

		msg := builder.Build()

		err = msg.Validate()
		if err != nil {
			return nil, err
		}

		response, err := t.Send(msg)
		if err != nil {
			return nil, err
		}

		// Parse response. If we got the "good" error, add the VLAN to the list.
		for _, a := range response.Attributes() {
			if a.Type() == attribute.ReplyStatusType {
				if a.String() == "Source Mac address not found" {
					found = append(found, vlan(i))
				}
			}
		}
	}
	return found, nil
}

func printResults(found []vlan) {
	fmt.Printf("\n%d VLANs found:", len(found))
	var somefound bool
	if len(found) > 0 {
		somefound = true
	}
	// Pretty print results
	var a []vlan
	// iterate over found VLAN numbers
	for _, v := range found {
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
				a = []vlan{v}
			}
		}
	}
	switch{
	case len(a) == 1:
		// Just one number? Print it.
		fmt.Printf(" %d", a[0])
	case len(a) > 1:
		// More than one numbers? Print as a range.
		fmt.Printf(" %d-%d", a[0], a[len(a)-1])
	}

	if somefound {
		fmt.Printf("\n")
	} else {
		fmt.Printf("<none>\n")
	}

}

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		log.Println("You need to specify a target switch")
		os.Exit(1)
	}

	t, err := target.TargetBuilder().
		AddIp(net.ParseIP(flag.Arg(0))).
		Build()
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	log.Println(t.String())

	found, err := enumerate_vlans(t)
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}

	printResults(found)
}
