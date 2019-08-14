package communicate

import (
	"log"
	"testing"
	"time"
)

// todo do some math here, check it out
func TestBackoffTicker(t *testing.T) {
	start := time.Now()
	end := start.Add(5*time.Second)
	bot := NewBackoffTicker(100*time.Millisecond)
	log.Println("start:",start)
	for time.Now().Before(end){
		select {
		case tick := <- bot.C:
			log.Println("tick:", tick)
		}
	}
	bot.Stop()
	time.Sleep(10*time.Second)
}

