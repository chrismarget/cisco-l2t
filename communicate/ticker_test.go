package communicate

import (
	"testing"
	"time"
)

func TestBackoffTicker(t *testing.T) {
	type testData struct {
		ticks           int
		initialInterval time.Duration
		estimate        time.Duration
	}

	var testDataSet []testData
	testDataSet = append(testDataSet,
		testData{
			ticks:           5,
			initialInterval: 100 * time.Millisecond,
			estimate:        800 * time.Millisecond,
		})
	testDataSet = append(testDataSet,
		testData{
			ticks:           8,
			initialInterval: 25 * time.Millisecond,
			estimate:        1600 * time.Millisecond,
		})

	for _, test := range testDataSet {
		start := time.Now()
		bot := NewBackoffTicker(test.initialInterval)
		threshold := test.estimate + (test.estimate / 20)
		end := start.Add(threshold)

		tickCount := 0
		for time.Now().Before(end) && tickCount < test.ticks {
			<-bot.C
			tickCount++
		}
		elapsed := time.Now().Sub(start)
		if elapsed >= threshold {
			t.Fatalf("test ran long: limit %s, elapsed %s", threshold, elapsed)
		}
		if elapsed < test.estimate {
			t.Fatalf("test ran short : limit %s, elapsed %s", test.estimate, elapsed)
		}
		bot.Stop()
	}
}
