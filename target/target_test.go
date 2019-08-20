package target

import (
	"github.com/chrismarget/cisco-l2t/attribute"
	"github.com/chrismarget/cisco-l2t/message"
	"log"
	"net"
	"testing"
	"time"
)

func TestNewTargetBuilder(t *testing.T) {
	tb, err := TargetBuilder().
		AddIp(net.ParseIP("192.168.8.254")).
		Build()
	if err != nil {
		t.Fatal(err)
	}

	println(tb.String())
}

func TestEstimateLatency(t *testing.T) {
	latency := []time.Duration{
		1 * time.Millisecond,
		2 * time.Millisecond,
		3 * time.Millisecond,
		4 * time.Millisecond,
		5 * time.Millisecond,
		6 * time.Millisecond,
		7 * time.Millisecond,
		8 * time.Millisecond,
		9 * time.Millisecond,
		10 * time.Millisecond,
		11 * time.Millisecond,
	}
	o := defaultTarget{
		info: []targetInfo{
			{
				rtt: latency,
			},
		},
		best: 0,
	}
	result := o.estimateLatency()
	if result.Round(100*time.Microsecond) == 11500*time.Microsecond {
		log.Printf("latency estimate okay: %s", result)
	} else {
		t.Fatalf("expected %d usec, got %d usec", 11500, result.Round(100*time.Microsecond))
	}

	if len(o.info[o.best].rtt) == 10 {
		log.Printf("latency samples %d - okay", len(o.info[o.best].rtt))
	} else {
		t.Fatalf("estimateLatency should truncate the latency slice to %d items", 10)
	}

	latency = []time.Duration{}
	o = defaultTarget{
		info: []targetInfo{
			{
				rtt: latency,
			},
		},
		best: 0,
	}
	result = o.estimateLatency()
	if result == 100*time.Millisecond {
		log.Printf("wild-ass guess latency okay: %s", result)
	} else {
		t.Fatalf("expected 100ms, got %s", result)
	}
}

func TestUpdateLatency(t *testing.T) {
	values := []time.Duration{
		4 * time.Millisecond,
		6 * time.Millisecond,
		2 * time.Millisecond,
		10 * time.Millisecond,
	}

	target := defaultTarget{
		info: []targetInfo{
			{rtt: nil},
		},
		best: 0,
	}

	for vi, val := range values {
		target.updateLatency(0, val)
		for i := 0; i <= vi; i++ {
			if values[i] != target.info[0].rtt[i] {
				t.Fatalf("target latency info not updating correctly")
			}
		}
	}
	log.Println("samples: ", len(target.info[0].rtt))
}

func TestGetVlans(t *testing.T) {
	target, err := TargetBuilder().
		AddIp(net.ParseIP("192.168.96.150")).
		Build()
	if err != nil {
		t.Fatal(err)
	}

	vlans, err := target.GetVlans()
	log.Println(vlans)

}

func TestSendBulkUnsafe(t *testing.T) {
	var bulkSendThis []message.Msg
	for i := 1; i <=10; i++{
		aSrcMac, err := attribute.NewAttrBuilder().
			SetType(attribute.SrcMacType).
			SetString("ffff.ffff.ffff").
			Build()
		if err != nil {
			t.Fatal(err)
		}
		aDstMac, err := attribute.NewAttrBuilder().
			SetType(attribute.DstMacType).
			SetString("ffff.ffff.ffff").
			Build()
		if err != nil {
			t.Fatal(err)
		}
		aVlan, err := attribute.NewAttrBuilder().
			SetType(attribute.VlanType).
			SetInt(uint32(i)).
			Build()
		if err != nil {
			t.Fatal(err)
		}
		aSrcIp, err := attribute.NewAttrBuilder().
			SetType(attribute.SrcIPv4Type).
			SetString("1.1.1.1").
			Build()
		if err != nil {
			t.Fatal(err)
		}
		msg := message.NewMsgBuilder().
			SetAttr(aSrcMac).
			SetAttr(aDstMac).
			SetAttr(aVlan).
			SetAttr(aSrcIp).
			Build()
		bulkSendThis = append(bulkSendThis, msg)
	}

	testTarget, err := TargetBuilder().
		AddIp(net.ParseIP("192.168.96.150")).
		Build()
	_ = testTarget
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(3*time.Second)

	result := testTarget.SendBulkUnsafe(bulkSendThis)
	log.Println(len(result))
}
