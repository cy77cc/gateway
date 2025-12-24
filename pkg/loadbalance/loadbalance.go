package loadbalance

import (
	"errors"

	"github.com/cy77cc/hioshop/common/register/types"
)

// LoadBalancer defines the interface for load balancing strategies
type LoadBalancer interface {
	Select(instances []*types.ServiceInstance) (*types.ServiceInstance, error)
}

var ErrNoInstances = errors.New("no instances available")
