package nacos

import (
	"fmt"

	"github.com/nacos-group/nacos-sdk-go/model"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

func (ins *Instance) GetAllServiceInstances(service string) ([]model.Instance, error) {
	// 这里是获取所有实例
	instances, err := ins.NamingClient.SelectInstances(vo.SelectInstancesParam{
		ServiceName: service,
		HealthyOnly: true,
	})
	if err != nil || len(instances) == 0 {
		return []model.Instance{}, fmt.Errorf("service %s not found", service)
	}

	return instances, nil
}
