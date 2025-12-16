package discovery

import (
	"fmt"
	"gateway-demo/config"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

type NacosClient struct {
	client naming_client.INamingClient
}

func NewNacosClient() *NacosClient {
	sc := []constant.ServerConfig{
		{
			IpAddr: config.CONFIG.Nacos.Addr,
			Port:   config.CONFIG.Nacos.Port,
		},
	}

	cc := constant.ClientConfig{
		NamespaceId: config.CONFIG.Nacos.Namespace,
		TimeoutMs:   config.CONFIG.Nacos.TimeoutMs,
		LogLevel:    "warn",
		AccessKey:   config.CONFIG.Nacos.AccessKey,
		SecretKey:   config.CONFIG.Nacos.SecretKey,
	}

	client, err := clients.NewNamingClient(vo.NacosClientParam{
		ServerConfigs: sc,
		ClientConfig:  &cc,
	})
	if err != nil {
		panic(err)
	}

	return &NacosClient{client: client}
}

func (n *NacosClient) Resolve(service string) (string, error) {
	// 这里是获取所有实例
	instances, err := n.client.SelectInstances(vo.SelectInstancesParam{
		ServiceName: service,
		HealthyOnly: true,
	})
	if err != nil || len(instances) == 0 {
		return "", fmt.Errorf("service %s not found", service)
	}

	ins := instances[0] // 第一阶段：先不做负载均衡
	return fmt.Sprintf("http://%s:%d", ins.Ip, ins.Port), nil
}
