package loadbalance

import (
	"errors"

	"github.com/cy77cc/gateway/pkg/discovery"
)

// LoadBalancer defines the interface for load balancing strategies
type LoadBalancer interface {
	Select(instances []discovery.Instance) (*discovery.Instance, error)
}

var ErrNoInstances = errors.New("no instances available")
