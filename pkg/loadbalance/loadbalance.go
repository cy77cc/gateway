package loadbalance

import (
	"errors"

	"github.com/cy77cc/hioshop/common/nacos"
)

// LoadBalancer defines the interface for load balancing strategies
type LoadBalancer interface {
	Select(instances []nacos.DiscoveryInstance) (*nacos.DiscoveryInstance, error)
}

var ErrNoInstances = errors.New("no instances available")
