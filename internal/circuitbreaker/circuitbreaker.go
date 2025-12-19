package circuitbreaker

import (
	"sync/atomic"
	"time"
)

type CircuitBreaker struct {
	state     int32
	failures  int64
	successes int64
	requests  int64
	openUntil int64
}

func (cb *CircuitBreaker) Allow() bool {
	state := atomic.LoadInt32(&cb.state)

	if state == Open {
		if time.Now().Unix() > atomic.LoadInt64(&cb.openUntil) {
			atomic.StoreInt32(&cb.state, HalfOpen)
			return true
		}
		return false
	}

	return true
}

func (cb *CircuitBreaker) OnResult(success bool, cfg CircuitConfig) {
	atomic.AddInt64(&cb.requests, 1)

	if success {
		atomic.AddInt64(&cb.successes, 1)
	} else {
		atomic.AddInt64(&cb.failures, 1)
	}

	req := atomic.LoadInt64(&cb.requests)
	if req < cfg.MinRequest {
		return
	}

	errRate := float64(atomic.LoadInt64(&cb.failures)) / float64(req)
	if errRate >= cfg.ErrorRate {
		atomic.StoreInt32(&cb.state, Open)
		atomic.StoreInt64(&cb.openUntil, time.Now().Add(time.Second*time.Duration(cfg.OpenSeconds)).Unix())
	}
}
