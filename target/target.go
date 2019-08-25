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
	GetLocalIp() net.IP
	HasIp(*net.IP) bool
	HasVlan(int) (bool, error)
	MacInVlan(net.HardwareAddr, int) (bool, error)
	Reachable() bool
	Send(message.Msg) (message.Msg, error)
	SendBulkUnsafe([]message.Msg, chan struct{}) []BulkSendResult
	SendUnsafe(message.Msg) communicate.SendResult
	String() string
}

type defaultTarget struct {
	reachable bool
	info      []targetInfo
	best      int
	name      string
	platform  string
	mgmtIp    net.IP
	rttLock   sync.Mutex
}

func (o *defaultTarget) GetLocalIp() net.IP {
	return o.info[o.best].localAddr
}

func (o *defaultTarget) Reachable() bool {
	return o.reachable
}

type BulkSendResult struct {
	Index   int
	Retries int
	Msg     message.Msg
	Err     error
}

func (o *defaultTarget) SendBulkUnsafe(out []message.Msg, progressChan chan struct{}) []BulkSendResult {
	resultChan := make(chan BulkSendResult, len(out))
	finalResultChan := make(chan []BulkSendResult)

	// Credit to stephen-fox for good ideas about using a buffered channel
	// to size a concurrency pool. Thank you Steve!
	//
	// Wait until all messages (len(out)) have been sent
	wg := &sync.WaitGroup{}
	wg.Add(len(out))

	// Initialize the worker pool (channel)
	maxWorkers := 100
	workers := 5
	workerPool := make(chan struct{}, maxWorkers)
	for i := 0; i < workers; i++ {
		workerPool <- struct{}{} //add worker credits to the workerPool
	}

	// collect results from all of the child routines;
	// tweak workerPool as necessary
	go func() {
		msgsSinceLastRetry := 0
		var results []BulkSendResult
		for r := range resultChan { // loop until resultChan closes
			results = append(results, r) // collect the reply
			wg.Done()

			updatePBar := true
			switch r.Err.(type) {
			case (net.Error):
				if r.Err.(net.Error).Temporary() {
					updatePBar = false
					msgsSinceLastRetry = 0
				}
				if r.Err.(net.Error).Timeout() {
					updatePBar = true
					msgsSinceLastRetry = 0
				}
			default:
				updatePBar = true
				if r.Retries == 0 {
					msgsSinceLastRetry++
				} else {
					msgsSinceLastRetry = 0
				}
			}

			if updatePBar {
				if progressChan != nil {
					progressChan <- struct{}{}
				}
			}

			switch msgsSinceLastRetry {
			case 0:
				if workers > 1 {
					<-workerPool
					workers--
				}
			case 5:
				if workers < maxWorkers {
					workerPool <- struct{}{}
					workers++
					msgsSinceLastRetry = 1
				}
			}
		}
		finalResultChan <- results
	}()

	// main loop instantiates a worker (pool permitting) to send each message
	for index, outMsg := range out {
		<-workerPool                    // Block until possible to get a worker credit
		go func(i int, m message.Msg) { // Start a worker routine
			reply := o.SendUnsafe(m)

			var inMsg message.Msg
			replyErr := reply.Err
			if replyErr == nil {
				inMsg, replyErr = message.UnmarshalMessageUnsafe(reply.ReplyData)
			}

			resultChan <- BulkSendResult{
				Index:   i,
				Retries: reply.Attempts - 1,
				Msg:     inMsg,
				Err:     replyErr,
			}
			workerPool <- struct{}{} // Worker done, return credit to the pool
		}(index, outMsg)

	}

	wg.Wait()
	close(resultChan)

	interimResults := <-finalResultChan // hope these are all good

	var goodResults []BulkSendResult
	var retry []message.Msg

	for _, ir := range interimResults {
		if x, ok := ir.Err.(net.Error); ok && x.Temporary() {
			retry = append(retry, out[ir.Index])
		} else {
			goodResults = append(goodResults, ir)
		}
	}

	var retryResult []BulkSendResult
	if len(retry) != 0 {
		retryResult = o.SendBulkUnsafe(retry, progressChan)
	}

	return append(goodResults, retryResult...)
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
		out.SetAttr(srcIpAttr)
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
		RttGuess:        o.estimateLatency(),
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

	out.WriteString("\n  Claimed IP:   ")
	switch o.mgmtIp.String() {
	case "<nil>":
		out.WriteString("<unknown>")
	default:
		out.WriteString(o.mgmtIp.String())
	}

	out.WriteString("\n  Known IP Addresses:")
	for _, i := range o.info {
		out.WriteString(fmt.Sprintf("\n    %15s responds from %-15s %s",
			i.destination.IP.String(),
			i.theirSource,
			averageRtt(i.rtt)))
	}

	out.WriteString("\n  Target address:      ")
	out.WriteString(o.info[o.best].destination.IP.String())

	out.WriteString("\n  Listen address:      ")
	out.WriteString(o.info[o.best].theirSource.String())

	out.WriteString("\n  Local address:       ")
	out.WriteString(o.info[o.best].localAddr.String())

	return out.String()
}

func averageRtt (in []time.Duration) time.Duration{
	var total time.Duration
	for _, i := range in {
		total = total + i
	}
	return total / time.Duration(len(in))
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
		observed = observed[lo-maxLatencySamples : lo]
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
