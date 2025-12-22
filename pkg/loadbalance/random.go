package loadbalance

import (
	"math/rand"
	"time"

	"github.com/cy77cc/hioshop/common/nacos"
)

type Random struct{}

func NewRandom() *Random {
	return &Random{}
}

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

func (r *Random) Select(instances []nacos.DiscoveryInstance) (*nacos.DiscoveryInstance, error) {
	if len(instances) == 0 {
		return nil, ErrNoInstances
	}

	index := rand.Intn(len(instances))
	return &instances[index], nil
}
