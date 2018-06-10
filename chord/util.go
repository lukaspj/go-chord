package chord

import (
	"time"
	"math"
	"github.com/lukaspj/go-logging/logging"
)

var (
	logger = logging.Logger{}
)

type tickingFunction struct {
	timer *time.Timer
	fn    func()
	stop  chan bool
	tick  chan bool
}

func StartTickingFunction(fn func() int) (tf tickingFunction) {
	tf.fn = func() {
		defer close(tf.stop)
		defer close(tf.tick)
		for {
			select {
			case <-tf.timer.C:
				go func() { tf.tick <- true }()
			case <-tf.tick:
				duration := fn()
				tf.timer.Reset(time.Duration(duration))
			case <-tf.stop:
				tf.timer.Stop()
				return
			}
		}
	}
	tf.timer = time.NewTimer(time.Second)
	tf.stop = make(chan bool)
	tf.tick = make(chan bool)

	go tf.fn()
	return
}

func cubic(min, max float64) (func(x float64) float64) {
	return func(x float64) float64 {
		mu := (1 - math.Cos(x*math.Pi)) / 2
		return min*(1-mu) + max*mu
	}
}