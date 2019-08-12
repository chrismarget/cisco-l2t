package rudp

import (
	"log"
	"math"
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
	ticker := time.NewTicker(d)
	stopChan := make(chan struct{})
	//wg := &sync.WaitGroup{}
	//wg.Add(1)

	//go func(c chan time.Time, t *time.Ticker, s chan struct{}, wg *sync.WaitGroup) {
	go func(c chan time.Time, t *time.Ticker, s chan struct{}) {
		var ticks float64 // these come at regular intervals
		var tocks float64 // these come at interval 0, 1,

		//c <- time.Now()
		//wg.Done()

		for mark := range t.C {
			select {
			case _, isOpen := <-stopChan:
				if !isOpen {
					log.Println("stopping")
					ticker.Stop()
					log.Println("stopped")
					return
				}
			default:
			}
			ticks++
			if math.Pow(2, tocks) == ticks {
				select {
				case c <- mark:
				default:
				}
				tocks++
			}
		}
		//	}(tockChan, ticker, stopChan, wg)
	}(tockChan, ticker, stopChan)

	//wg.Wait()

	return &BackoffTicker{
		C:        tockChan,
		stopChan: stopChan,
	}
}
