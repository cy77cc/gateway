package nacos

import (
	"gateway/config"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

type Instance struct {
	clientConfig *constant.ClientConfig
	serverConfig []constant.ServerConfig
	ConfigClient config_client.IConfigClient
	NamingClient naming_client.INamingClient
}

func NewNacosInstance(cfg config.NacosConfig) (*Instance, error) {
	sc := []constant.ServerConfig{
		{
			IpAddr:      cfg.Endpoint,
			Port:        cfg.Port,
			ContextPath: cfg.ContextPath,
		},
	}

	cc := &constant.ClientConfig{
		NamespaceId: cfg.Namespace,
		TimeoutMs:   cfg.TimeoutMs,
		LogLevel:    "warn",
		Username:    cfg.Username,
		Password:    cfg.Password,
		ContextPath: cfg.ContextPath,
		Endpoint:    cfg.Endpoint,
	}

	namingClient, err := clients.NewNamingClient(vo.NacosClientParam{
		ServerConfigs: sc,
		ClientConfig:  cc,
	})
	if err != nil {
		return nil, err
	}

	configClient, err := clients.NewConfigClient(vo.NacosClientParam{
		ServerConfigs: sc,
		ClientConfig:  cc,
	})

	if err != nil {
		return nil, err
	}

	return &Instance{
		clientConfig: cc,
		serverConfig: sc,
		ConfigClient: configClient,
		NamingClient: namingClient,
	}, nil

}
