package runy

import (
	"os"
	"os/signal"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSetupSignalHandler(t *testing.T) {
	ctx := SetupSignalHandler()
	task := &Task{ticker: time.NewTicker(time.Second * 2)}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	task.wg.Add(1)
	go func(c chan os.Signal) {
		defer task.wg.Done()
		task.Run(c)
	}(c)

	select {
	case <-c:
		return
	case _, ok := <-ctx.Done():
		assert.False(t, ok)
	}
}

type Task struct {
	wg     sync.WaitGroup
	ticker *time.Ticker
}

func (t *Task) Run(c chan os.Signal) {
	for {
		go sendSignal(c)
		handle()
	}
}

func handle() {
	for i := 0; i < 5; i++ {
		time.Sleep(time.Millisecond * 100)
	}
}

func sendSignal(stopChan chan os.Signal) {
	time.Sleep(1 * time.Second)
	stopChan <- os.Interrupt
}
