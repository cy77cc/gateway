package nacos

import (
	"fmt"
	"gateway/pkg/discovery"

	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// GetService implements discovery.ServiceDiscovery
func (ins *Instance) GetService(serviceName string) ([]discovery.Instance, error) {
	instances, err := ins.NamingClient.SelectInstances(vo.SelectInstancesParam{
		ServiceName: serviceName,
		HealthyOnly: true,
	})
	if err != nil {
		return nil, err
	}

	if len(instances) == 0 {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}

	var result []discovery.Instance
	for _, inst := range instances {
		result = append(result, discovery.Instance{
			ID:       inst.InstanceId,
			Host:     inst.Ip,
			Port:     int(inst.Port),
			Metadata: inst.Metadata,
			Weight:   inst.Weight,
		})
	}
	return result, nil
}
