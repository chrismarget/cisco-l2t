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
	"sort"
)

const (
	vlanMin = 1
	vlanMax = 4094
)

func queryTemplate(t target.Target) (message.Msg, error) {
	srcIpAttr, err := attribute.NewAttrBuilder().
		SetType(attribute.SrcIPv4Type).
		SetString(t.GetLocalIp().String()).
		Build()
	if err != nil {
		return nil, err
	}

	msg, err := message.TestMsg()
	if err != nil {
		return nil, err
	}

	msg.SetAttr(srcIpAttr)
	return msg, nil
}

func buildQueries(t target.Target) ([]message.Msg, error) {
	var queries []message.Msg
	for v := vlanMin; v <= vlanMax; v++ {
		template, err := queryTemplate(t)
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

		template.SetAttr(vlanAttr)
		queries = append(queries, template)
	}
	return queries, nil
}

func enumerate_vlans(t target.Target) ([]int, error) {
	queries, err := buildQueries(t)
	if err != nil {
		return nil, err
	}

	var found []int
	for len(queries) > 0 {
		// progress bar and bar channel
		bar := pb.StartNew(len(queries))
		pChan := make(chan struct{})
		go func() {
			for _ = range pChan {
				bar.Increment()
			}
			bar.Finish()
		}()

		// go do work
		result := t.SendBulkUnsafe(queries, pChan)

		// how'd we do?
		var doOver []message.Msg
		for i, r := range result {
			if r.Err != nil {
				// Error? This message goes on the doOver slice
				doOver = append(doOver, queries[i])
				continue
			}
			err := r.Msg.Validate()
			if err != nil {
				// Message validation error? This message goes on the doOver slice
				doOver = append(doOver, queries[i])
				continue
			}
			replyStatus := r.Msg.GetAttr(attribute.ReplyStatusType)
			if replyStatus == nil {
				// No reply status attribute? This message goes on the doOver slice
				doOver = append(doOver, queries[i])
				continue
			}
			// So far so good. Check to see whether the replyStatus indicates the vlan exists
			if replyStatus.String() == attribute.ReplyStatusSrcNotFound {
				found = append(found, r.Index+vlanMin)
			}
		}
		// if there's any doOver messages this will send us back around for them
		queries = doOver
		if len(queries) > 0 {
			fmt.Println("missed some...")
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
	fmt.Print(t.String(), "\n")

	found, err := enumerate_vlans(t)
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}

	printResults(found)
}
