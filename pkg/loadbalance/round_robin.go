package loadbalance

import (
	"sync/atomic"

	"github.com/cy77cc/hioshop/common/register/types"
)

type RoundRobin struct {
	counter uint64
}

func NewRoundRobin() *RoundRobin {
	return &RoundRobin{}
}

func (rb *RoundRobin) Select(instances []*types.ServiceInstance) (*types.ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, ErrNoInstances
	}

	count := atomic.AddUint64(&rb.counter, 1)
	index := (count - 1) % uint64(len(instances))
	return instances[index], nil
}
