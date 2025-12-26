package circuitbreaker

import (
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cy77cc/gateway/config"
)

const (
	Closed int32 = iota
	Open
	HalfOpen
)

// Bucket 时间窗口桶
type Bucket struct {
	successes int64
	failures  int64
	timestamp int64
}

// SlidingWindow 滑动时间窗口
type SlidingWindow struct {
	buckets []*Bucket
	size    int
	index   int
	mutex   sync.Mutex
}

// NewSlidingWindow 创建新的滑动窗口
func NewSlidingWindow(size int) *SlidingWindow {
	buckets := make([]*Bucket, size)
	for i := range buckets {
		buckets[i] = &Bucket{timestamp: time.Now().Unix()}
	}
	return &SlidingWindow{
		buckets: buckets,
		size:    size,
		index:   0,
	}
}

// Add 添加请求结果
func (sw *SlidingWindow) Add(success bool) {
	sw.mutex.Lock()
	defer sw.mutex.Unlock()

	now := time.Now().Unix()
	currentBucket := sw.buckets[sw.index]

	// 如果是新的时间单位，移动到下一个桶
	if now > currentBucket.timestamp {
		sw.index = (sw.index + 1) % sw.size
		currentBucket = sw.buckets[sw.index]
		currentBucket.successes = 0
		currentBucket.failures = 0
		currentBucket.timestamp = now
	}

	if success {
		atomic.AddInt64(&currentBucket.successes, 1)
	} else {
		atomic.AddInt64(&currentBucket.failures, 1)
	}
}

// GetMetrics 获取窗口内的统计数据
func (sw *SlidingWindow) GetMetrics() (total, failures int64) {
	sw.mutex.Lock()
	defer sw.mutex.Unlock()

	now := time.Now().Unix()
	var totalRequests, failedRequests int64

	for _, bucket := range sw.buckets {
		// 只统计最近一段时间内的数据
		if now-bucket.timestamp < int64(sw.size) {
			totalRequests += atomic.LoadInt64(&bucket.successes)
			totalRequests += atomic.LoadInt64(&bucket.failures)
			failedRequests += atomic.LoadInt64(&bucket.failures)
		}
	}

	return totalRequests, failedRequests
}

type CircuitBreaker struct {
	state                int32
	window               *SlidingWindow
	openUntil            int64
	halfOpenCh           chan struct{}
	consecutiveSuccesses int64
}

// NewCircuitBreaker 创建新的熔断器
func NewCircuitBreaker(windowSize int) *CircuitBreaker {
	cb := &CircuitBreaker{
		state:      Closed,
		window:     NewSlidingWindow(windowSize),
		halfOpenCh: make(chan struct{}, 1),
	}
	return cb
}

// Allow 判断是否允许请求通过
func (cb *CircuitBreaker) Allow() bool {
	state := atomic.LoadInt32(&cb.state)

	switch state {
	case Open:
		if time.Now().Unix() > atomic.LoadInt64(&cb.openUntil) {
			// 尝试进入半开状态
			if atomic.CompareAndSwapInt32(&cb.state, Open, HalfOpen) {
				atomic.StoreInt64(&cb.consecutiveSuccesses, 0)
				return true
			}
		}
		return false
	case HalfOpen:
		// 半开状态下限制并发请求
		select {
		case cb.halfOpenCh <- struct{}{}:
			return true
		default:
			return false
		}
	default: // Closed
		return true
	}
}

// calculateK 计算K值，基于请求量动态调整
func (cb *CircuitBreaker) calculateK(requests int64) float64 {
	// K值随着请求量增加而减小，提高敏感度
	if requests < 10 {
		return 2.0
	} else if requests < 100 {
		return 1.5
	}
	return 1.2
}

// shouldTrip 判断是否应该触发熔断
func (cb *CircuitBreaker) shouldTrip(total, failures int64, cfg *config.CircuitConfig) bool {
	if total < cfg.MinRequest {
		return false
	}

	// SRE弹性熔断算法公式
	// P_e = failures/total (错误概率)
	// 当 P_e > (P_e_max * (1 + K*sqrt((1-P_e_max)/total)))

	P_e := float64(failures) / float64(total)
	P_e_max := cfg.ErrorRate

	K := cb.calculateK(total)
	threshold := P_e_max * (1 + K*math.Sqrt((1-P_e_max)/float64(total)))

	return P_e > threshold
}

// OnResult 处理请求结果
func (cb *CircuitBreaker) OnResult(success bool, cfg *config.CircuitConfig) {
	// 释放半开状态的令牌
	if atomic.LoadInt32(&cb.state) == HalfOpen {
		select {
		case <-cb.halfOpenCh:
		default:
		}
	}

	// 记录请求结果
	cb.window.Add(success)

	// 获取统计信息
	total, failures := cb.window.GetMetrics()

	state := atomic.LoadInt32(&cb.state)

	switch state {
	case Closed:
		// 在关闭状态下检查是否需要熔断
		if cb.shouldTrip(total, failures, cfg) {
			atomic.StoreInt32(&cb.state, Open)
			atomic.StoreInt64(&cb.openUntil, time.Now().Add(time.Second*time.Duration(cfg.OpenSeconds)).Unix())
		}
	case HalfOpen:
		if success {
			atomic.AddInt64(&cb.consecutiveSuccesses, 1)
			// 连续成功次数达到阈值，恢复到关闭状态
			if atomic.LoadInt64(&cb.consecutiveSuccesses) >= cfg.HalfOpenSuccess {
				atomic.StoreInt32(&cb.state, Closed)
				atomic.StoreInt64(&cb.consecutiveSuccesses, 0)
			}
		} else {
			// 半开状态下失败，重新进入熔断状态
			atomic.StoreInt32(&cb.state, Open)
			atomic.StoreInt64(&cb.openUntil, time.Now().Add(time.Second*time.Duration(cfg.OpenSeconds)).Unix())
		}
	case Open:
		// 开启状态下不需要特殊处理
	}
}
