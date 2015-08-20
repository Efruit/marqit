package exchange

import (
	"github.com/inconshreveable/log15"
	"math/rand"
	"strconv"
	"time"
)

type Timer struct {
	C         chan time.Time
	stop      chan struct{}
	pause     chan struct{}
	resume    chan struct{}
	startedat time.Time
	pausedat  time.Time
	remaining time.Duration
	finished  bool
	number    uint64
}

func (t *Timer) Stop() {
	if t.finished {
		panic("Stop on finished timer. n=" + strconv.FormatUint(t.number, 10))
	}
	t.stop <- struct{}{}
}

func (t *Timer) Pause() {
	if t.finished {
		panic("Pause on finished timer. n=" + strconv.FormatUint(t.number, 10))
	}
	t.pause <- struct{}{}
}

func (t *Timer) Resume() {
	if t.finished {
		panic("Resume on finished timer. n=" + strconv.FormatUint(t.number, 10))
	}
	t.resume <- struct{}{}
}

func (t *Timer) runTimer() {
	log15.Info("Timer started", "number", t.number, "remaining", t.remaining)
	tt := time.NewTimer(t.remaining)
	for {
		select {
		case <-t.stop:
			log15.Info("Stopping timer", "number", t.number, "remaining", t.remaining)
			t.finished = true
			close(t.C)
			return
		case <-t.pause:
			log15.Info("Pausing timer", "number", t.number, "remaining", t.remaining)
			t.pausedat = time.Now()
			t.remaining -= t.pausedat.Sub(t.startedat)
			tt.Stop()
		case <-t.resume:
			t.startedat = time.Now()
			elapsed := t.startedat.Sub(t.pausedat)
			if elapsed > t.remaining {
				log15.Info("Timer completed (b)", "number", t.number, "remaining", t.remaining)
				t.C <- time.Now().Add(-t.remaining)
				t.finished = true
				return
			}
			log15.Info("Resuming timer", "number", t.number, "remaining", t.remaining, "elapsed", elapsed)
			tt = time.NewTimer(t.remaining - elapsed)
			t.remaining -= elapsed
		case ttt := <-tt.C:
			log15.Info("Timer completed", "number", t.number, "remaining", t.remaining)
			t.C <- ttt
			t.finished = true
			return
		}
	}
}

func NewTimer(tt time.Duration) *Timer {
	t := &Timer{
		C:         make(chan time.Time, 1),
		stop:      make(chan struct{}),
		pause:     make(chan struct{}),
		resume:    make(chan struct{}),
		startedat: time.Now(),
		remaining: tt,
		number:    uint64(rand.Uint32()),
	}
	go t.runTimer()
	return t
}
