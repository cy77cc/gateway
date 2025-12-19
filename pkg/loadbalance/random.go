package loadbalance

import (
	"github.com/cy77cc/gateway/pkg/discovery"
	"math/rand"
	"time"
)

type Random struct{}

func NewRandom() *Random {
	return &Random{}
}

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

func (r *Random) Select(instances []discovery.Instance) (*discovery.Instance, error) {
	if len(instances) == 0 {
		return nil, ErrNoInstances
	}

	index := rand.Intn(len(instances))
	return &instances[index], nil
}
