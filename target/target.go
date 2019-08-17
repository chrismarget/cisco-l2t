package target

import (
	"bytes"
	"fmt"
	"github.com/chrismarget/cisco-l2t/attribute"
	"github.com/chrismarget/cisco-l2t/communicate"
	"github.com/chrismarget/cisco-l2t/message"
	"net"
	"time"
)

const (
	maxLatencySamples = 10
)

type Target interface {
	GetVlans() ([]int, error)
	HasIp(*net.IP) bool
	HasVlan(int) (bool, error)
	Send(message.Msg) (message.Msg, error)
	SendUnsafe(message.Msg) (message.Msg, error)
	String() string
}

type defaultTarget struct {
	info     []targetInfo
	best     int
	name     string
	platform string
}

func (o *defaultTarget) Send(out message.Msg) (message.Msg, error) {
	in, err := o.SendUnsafe(out)
	if err != nil {
		return nil, err
	}

	err = in.Validate()
	if err != nil {
		return in, err
	}

	return in, nil
}

func (o *defaultTarget) SendUnsafe(msg message.Msg) (message.Msg, error) {
	var payload []byte
	switch msg.NeedsSrcIp() {
	case true:
		srcIpAttr, err := attribute.NewAttrBuilder().
			SetType(attribute.SrcIPv4Type).
			SetString(o.info[o.best].localAddr.String()).
			Build()
		if err != nil {
			return nil, err
		}
		payload = msg.Marshal([]attribute.Attribute{srcIpAttr})
	case false:
		payload = msg.Marshal([]attribute.Attribute{})
	}

	out := communicate.SendThis{
		Payload:         payload,
		Destination:     o.info[o.best].destination,
		ExpectReplyFrom: o.info[o.best].theirSource,
		RttGuess:        communicate.InitialRTTGuess,
	}

	in := communicate.Communicate(out, nil)

	if in.Err != nil {
		return nil, in.Err
	}

	return message.UnmarshalMessageUnsafe(in.ReplyData)
}

func (o *defaultTarget) String() string {
	var out bytes.Buffer

	out.WriteString("Target info:\n  Hostname:     ")
	switch o.name {
	case "":
		out.WriteString("<unknown>")
	default:
		out.WriteString(o.name)
	}

	out.WriteString("\n  Platform:     ")
	switch o.platform {
	case "":
		out.WriteString("<unknown>")
	default:
		out.WriteString(o.platform)
	}

	out.WriteString("\n  Known IP Addresses:")
	for _, i := range o.info {
		out.WriteString(fmt.Sprintf("\n    %15s responds from %-15s %s",
			i.destination.IP.String(),
			i.theirSource,
			i.rtt))
	}

	out.WriteString("\n  Target address:      ")
	out.WriteString(o.info[o.best].destination.IP.String())

	out.WriteString("\n  Listen address:      ")
	out.WriteString(o.info[o.best].theirSource.String())

	out.WriteString("\n  Local address:       ")
	out.WriteString(o.info[o.best].localAddr.String())

	return out.String()
}

// estimateLatency tries to estimate the response time for this target
// using the contents of the objects latency slice.
func (o *defaultTarget) estimateLatency() time.Duration {
	observed := o.info[o.best].rtt
	if len(observed) == 0 {
		return communicate.InitialRTTGuess
	}

	// trim the latency samples
	if len(observed) > maxLatencySamples {
		o.info[o.best].rtt = observed[:maxLatencySamples]
	}

	// half-assed latency estimator does a rolling average then pads 25%
	var result int64
	for i, l := range observed {
		switch i {
		case 0:
			result = int64(l)
		default:
			result = (result + int64(l)) / 2
		}
	}
	return time.Duration(float32(result) * float32(1.25))
}

// updateLatency adds the passed time.Duration as the most recent
// latency sample to the specified targetInfo index.
func (o *defaultTarget) updateLatency(index int, t time.Duration) {
	upperBound := maxLatencySamples
	if len(o.info[index].rtt) < maxLatencySamples-1 {
		// not many samples here. set the upper bound (used below)
		// to match the sample count after the append()
		upperBound = len(o.info[index].rtt) + 1
	}
	o.info[index].rtt = append(o.info[index].rtt, t)[:upperBound]
}

type SendMessageConfig struct {
	M     message.Msg
	Inbox chan MessageResponse
}

type MessageResponse struct {
	Response message.Msg
	Err      error
}
