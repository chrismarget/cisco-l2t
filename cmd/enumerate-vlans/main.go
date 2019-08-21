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
	"sort"
)

const (
	vlanMin = 1
	vlanMax = 4094
)

func enumerate_vlans(t target.Target) ([]int, error) {
	//bar := pb.StartNew(vlanMax)

	var queries []message.Msg

	srcIpAttr, err := attribute.NewAttrBuilder().
		SetType(attribute.SrcIPv4Type).
		SetString(t.GetLocalIp().String()).
		Build()
	if err != nil {
		return nil, err
	}

	// loop over all VLANs
	for v := vlanMin; v <= vlanMax; v++ {
		msg, err := message.TestMsg()
		if err != nil {
			return nil, err
		}

		vlanAttr, err := attribute.NewAttrBuilder().
			SetType(attribute.VlanType).
			SetInt(uint32(v)).
			Build()
		if err != nil {
			return nil, err
		}

		msg.SetAttr(srcIpAttr)
		msg.SetAttr(vlanAttr)

		queries = append(queries, msg)

		//bar.Increment()
		//vlanFound, err := t.HasVlan(v)
		//if err != nil {
		//	return found, err
		//}
		//if vlanFound {
		//	found = append(found, v)
		//}
	}

	result := t.SendBulkUnsafe(queries)
	var found []int
	for _, r := range result {
		for _, a := range r.Msg.Attributes() {
			if a.Type() == attribute.ReplyStatusType &&
				a.String() == attribute.ReplyStatusSrcNotFound {
				found = append(found, r.Index)
			}
		}
	}

	return found, nil
}

func printResults(found []int) {
	sort.Ints(found)
	fmt.Printf("\n%d VLANs found:", len(found))
	var somefound bool
	if len(found) > 0 {
		somefound = true
	}
	// Pretty print results
	var a []int
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
				a = []int{v}
			}
		}
	}
	switch {
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

	fmt.Print(t.String(),"\n")

	found, err := enumerate_vlans(t)
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}

	printResults(found)
}
