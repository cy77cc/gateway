package nacos

import (
	"gateway-demo/config"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

type Instance struct {
	clientConfig *constant.ClientConfig
	serverConfig []constant.ServerConfig
	ConfigClient config_client.IConfigClient
	NamingClient naming_client.INamingClient
}

func NewNacosInstance(cfg *config.NacosConfig) *Instance {
	sc := []constant.ServerConfig{
		{
			IpAddr: cfg.Endpoint,
			Port:   cfg.Port,
		},
	}

	cc := &constant.ClientConfig{
		NamespaceId: cfg.Namespace,
		TimeoutMs:   cfg.TimeoutMs,
		LogLevel:    "warn",
		Username:    cfg.Username,
		Password:    cfg.Password,
		ContextPath: cfg.ContextPath,
	}

	namingClient, err := clients.NewNamingClient(vo.NacosClientParam{
		ServerConfigs: sc,
		ClientConfig:  cc,
	})
	if err != nil {
		panic(err)
	}

	configClient, err := clients.NewConfigClient(vo.NacosClientParam{
		ServerConfigs: sc,
		ClientConfig:  cc,
	})

	if err != nil {
		panic(err)
	}

	return &Instance{
		clientConfig: cc,
		serverConfig: sc,
		ConfigClient: configClient,
		NamingClient: namingClient,
	}

}
