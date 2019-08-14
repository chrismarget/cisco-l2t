package rudp

import (
	"math"
	"sync"
	"time"
)

type BackoffTicker struct {
	ticker   time.Ticker
	C        chan time.Time
	stopChan chan struct{}
}

func (o *BackoffTicker) Stop() {
	close(o.stopChan)
}

func NewBackoffTicker(d time.Duration) *BackoffTicker {
	tockChan := make(chan time.Time, 1)
	tockChan <- time.Now()
	wg := &sync.WaitGroup{}
	wg.Add(1)
	stopChan := make(chan struct{})

	go func() {
		wg.Done()
		ticker := time.NewTicker(d)
		var ticks float64 // these come at regular intervals
		var tocks float64 // these come at interval 0, 1,

		for mark := range ticker.C {
			select {
			case _, isOpen := <-stopChan:
				if !isOpen {
					ticker.Stop()
					return
				}
			default:
			}
			ticks++
			if math.Pow(2, tocks) == ticks {
				select {
				case tockChan <- mark:
				default:
				}
				tocks++
			}
		}
	}()

	wg.Wait()

	return &BackoffTicker{
		C:        tockChan,
		stopChan: stopChan,
	}
}
