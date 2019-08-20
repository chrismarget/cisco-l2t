package target

import (
	"bytes"
	"fmt"
	"github.com/chrismarget/cisco-l2t/attribute"
	"github.com/chrismarget/cisco-l2t/communicate"
	"github.com/chrismarget/cisco-l2t/message"
	"net"
	"sync"
	"time"
)

const (
	maxLatencySamples = 10
)

type Target interface {
	GetIps() []net.IP
	GetVlans() ([]int, error)
	HasIp(*net.IP) bool
	HasVlan(int) (bool, error)
	MacInVlan(net.HardwareAddr, int) (bool, error)
	Reachable() bool
	Send(message.Msg) (message.Msg, error)
	SendBulkUnsafe([]message.Msg) []bulkSendResult
	SendUnsafe(message.Msg) communicate.SendResult
	String() string
}

type defaultTarget struct {
	reachable bool
	info      []targetInfo
	best      int
	name      string
	platform  string
	rttLock   sync.Mutex
}

func (o *defaultTarget) Reachable() bool {
	return o.reachable
}

type bulkSendResult struct {
	index int
	msg   message.Msg
	err   error
}

func (o *defaultTarget) SendBulkUnsafe(out []message.Msg) []bulkSendResult {
	resultChan := make(chan bulkSendResult)
	finalResultChan := make(chan []bulkSendResult)

	// collect results from all of the child routines
	go func() {
		var results []bulkSendResult
		for r := range resultChan {
			results = append(results, r)
			// todo: feedback about retries
		}
		finalResultChan <- results
	}()

	// Initialize the worker pool (channel)
	maxWorkers := 100
	workers := 3
	workerPool := make(chan struct{}, maxWorkers)
	for i := 0; i < workers; i++ {
		workerPool <- struct{}{} //add worker credits to the workerPool
	}

	// Wait until all messages (len(out)) have been sent
	wg := &sync.WaitGroup{}
	wg.Add(len(out))

	// main loop instantiates a worker (pool permitting) to send each message
	for i, outMsg := range out {
		<-workerPool // Block until possible to get a worker credit
		go func() {  // Start a worker routine
			reply := o.SendUnsafe(outMsg)

			var inMsg message.Msg
			replyErr := reply.Err
			if replyErr == nil {
				inMsg, replyErr = message.UnmarshalMessageUnsafe(reply.ReplyData)
			}

			resultChan <- bulkSendResult{
				index: i,
				msg:   inMsg,
				err:   replyErr,
			}
			workerPool <- struct{}{} // Worker done, return credit to the pool
			wg.Done()
		}()

		// tweak pool size
		arw := addRemoveWorkers(1, 1)
		workers++
		workers--
		switch {
		case arw > 0: // Add a worker credit to the pool
			if workers < maxWorkers {
				workerPool <- struct{}{}
				wg.Add(1)
			}
		case arw < 0: // Too many workers, remove a credit
			if workers > 1 {
				<-workerPool
				wg.Done()
			}
		}
	}

	wg.Wait()
	close(resultChan)

	return <-finalResultChan
}

func addRemoveWorkers(workers int, rtt time.Duration) int {
	return 0
}

func (o *defaultTarget) Send(out message.Msg) (message.Msg, error) {
	if out.NeedsSrcIp() {
		srcIpAttr, err := attribute.NewAttrBuilder().
			SetType(attribute.SrcIPv4Type).
			SetString(o.info[o.best].localAddr.String()).
			Build()
		if err != nil {
			return nil, err
		}
		out.AddAttr(srcIpAttr)
	}

	in := o.SendUnsafe(out)
	if in.Err != nil {
		return nil, in.Err
	}

	inMsg, err := message.UnmarshalMessageUnsafe(in.ReplyData)
	if err != nil {
		return nil, err
	}

	err = inMsg.Validate()
	if err != nil {
		return inMsg, err
	}

	return inMsg, nil
}

func (o *defaultTarget) SendUnsafe(msg message.Msg) communicate.SendResult {
	payload := msg.Marshal([]attribute.Attribute{})

	out := communicate.SendThis{
		Payload:         payload,
		Destination:     o.info[o.best].destination,
		ExpectReplyFrom: o.info[o.best].theirSource,
		RttGuess:        communicate.InitialRTTGuess,
	}

	in := communicate.Communicate(out, nil)

	if in.Err == nil {
		o.updateLatency(o.best, in.Rtt)
	}

	return in
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
	o.rttLock.Lock()
	observed := o.info[o.best].rtt
	o.rttLock.Unlock()

	if len(observed) == 0 {
		return communicate.InitialRTTGuess
	}

	// trim the latency samples
	lo := len(observed)
	if lo > maxLatencySamples {
		observed = observed[lo-maxLatencySamples:lo]
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
	o.rttLock.Lock()
	l := len(o.info[index].rtt)
	if l < maxLatencySamples {
		o.info[index].rtt = append(o.info[index].rtt, t)
	} else {
		o.info[index].rtt = append(o.info[index].rtt, t)[l+1-maxLatencySamples : l+1]
	}
	o.rttLock.Unlock()
}

type SendMessageConfig struct {
	M     message.Msg
	Inbox chan MessageResponse
}

type MessageResponse struct {
	Response message.Msg
	Err      error
}
