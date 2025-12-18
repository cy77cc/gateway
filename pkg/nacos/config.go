package nacos

import (
	"fmt"
	"gateway/config"
	"log"

	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// LoadAndWatchConfig 从Nacos加载配置并监听配置变化
func (ins *Instance) LoadAndWatchConfig(dataId, group string) error {
	// 从Nacos获取配置
	content, err := ins.ConfigClient.GetConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
	})
	if err != nil {
		return fmt.Errorf("failed to get config from nacos: %w", err)
	}

	// 应用初始配置
	if err := applyConfig(content, dataId); err != nil {
		return fmt.Errorf("failed to apply initial config: %w", err)
	}

	// 设置配置变更回调函数
	onChangeCallback := func(namespace, group, dataId, data string) {
		log.Printf("config changed - namespace: %s, group: %s, dataId: %s", namespace, group, dataId)

		// 应用新的配置
		if err := applyConfig(data, dataId); err != nil {
			log.Printf("failed to apply changed config: %v", err)
			return
		}
	}

	// 监听配置变化
	err = ins.ConfigClient.ListenConfig(vo.ConfigParam{
		DataId:   dataId,
		Group:    group,
		OnChange: onChangeCallback,
	})
	if err != nil {
		return fmt.Errorf("failed to listen config changes: %w", err)
	}

	log.Printf("successfully loaded and watching config - dataId: %s, group: %s", dataId, group)
	return nil
}

// applyConfig 应用配置内容
func applyConfig(content string, dataId string) error {
	log.Printf("applying config for %s", dataId)

	switch dataId {
	case "gateway-router":
		return config.UpdateRouteConfig(content)
	case "gateway-global":
		return config.UpdateServerConfig(content)
	default:
		return fmt.Errorf("unknown config dataId: %s", dataId)
	}
}
