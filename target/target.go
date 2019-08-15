package target

import (
	"bytes"
	"github.com/chrismarget/cisco-l2t/attribute"
	"github.com/chrismarget/cisco-l2t/communicate"
	"github.com/chrismarget/cisco-l2t/message"
	"net"
	"strconv"
	"time"
)

const (
	nilIP             = "<nil>"
	maxLatencySamples = 10
)

type Target interface {
	Send(message.Msg) (message.Msg, error)
	String() string
}

type targetInfo struct{
	theirSource net.IP
	rtt []time.Duration
}

type defaultTarget struct {
	destination     *net.UDPAddr
	theirIp         []net.IP
	talkToThemIdx   int
	listenToThemIdx int
	ourIp           net.IP
	useDial         bool
	latency         []time.Duration
	name            string
	platform        string
}

func (o *defaultTarget) Send(msg message.Msg) (message.Msg, error) {
	var payload []byte
	switch msg.NeedsSrcIp() {
	case true:
		srcIpAttr, err := attribute.NewAttrBuilder().
			SetType(attribute.SrcIPv4Type).
			SetString(o.ourIp.String()).
			Build()
		if err != nil {
			return nil, err
		}
		payload = msg.Marshal([]attribute.Attribute{srcIpAttr})
	case false:
		payload = msg.Marshal([]attribute.Attribute{})
	}

	out := communicate.SendThis{
		Payload: payload,
		Destination: &net.UDPAddr{
			IP:   o.theirIp[o.talkToThemIdx],
			Port: communicate.CiscoL2TPort,
		},
		ExpectReplyFrom: o.theirIp[o.listenToThemIdx],
		RttGuess:        communicate.InitialRTTGuess,
	}

	in := communicate.Communicate(out, nil)

	if in.Err != nil {
		return nil, in.Err
	}

	return message.UnmarshalMessage(in.ReplyData)
}

func (o *defaultTarget) String() string {
	var out bytes.Buffer

	out.WriteString("\nTarget Hostname:    ")
	switch o.name {
	case "":
		out.WriteString("<unknown>")
	default:
		out.WriteString(o.name)
	}

	out.WriteString("\nTarget Platform:    ")
	switch o.platform {
	case "":
		out.WriteString("<unknown>")
	default:
		out.WriteString(o.platform)
	}

	out.WriteString("\nKnown IP Addresses:")
	for _, ip := range o.theirIp {
		out.WriteString(" ")
		out.WriteString(ip.String())
	}

	out.WriteString("\nTarget address:     ")
	switch {
	case o.talkToThemIdx >= 0:
		out.WriteString(o.theirIp[o.talkToThemIdx].String())
	default:
		out.WriteString("none")
	}

	out.WriteString("\nListen address:     ")
	switch {
	case o.listenToThemIdx >= 0:
		out.WriteString(o.theirIp[o.listenToThemIdx].String())
	default:
		out.WriteString("none")
	}

	out.WriteString("\nLocal address:      ")
	switch o.ourIp.String() {
	case nilIP:
		out.WriteString("none")
	default:
		out.WriteString(o.ourIp.String())
	}

	out.WriteString("\nUse Dial:           ")
	out.WriteString(strconv.FormatBool(o.useDial))

	return out.String()
}

// estimateLatency tries to estimate the response time for this target
// using the contents of the objects latency slice.
func (o *defaultTarget) estimateLatency() time.Duration {
	if len(o.latency) == 0 {
		return communicate.InitialRTTGuess
	}
	var result int64
	for i, l := range o.latency {
		switch i {
		case 0:
			result = int64(l)
		default:
			result = (result + int64(l)) / 2
		}
	}
	return time.Duration(float32(result) * float32(1.15))
}

// updateLatency adds the passed time.Duration as the most recent
// latency sample, trims the latency slice to size.
func (o *defaultTarget) updateLatency(t time.Duration) {
	o.latency = append(o.latency, t)
	// delete old elements of latency slice because we care more about
	// recent data (and certainly want to purge early bad assumptions)
	if len(o.latency) > maxLatencySamples {
		o.latency = o.latency[len(o.latency)-maxLatencySamples : len(o.latency)]
	}

}

type SendMessageConfig struct {
	M     message.Msg
	Inbox chan MessageResponse
}

type MessageResponse struct {
	Response message.Msg
	Err      error
}
