package nacos

import (
	"fmt"
	"gateway/config"
	"log"

	"github.com/nacos-group/nacos-sdk-go/vo"
)

// LoadAndWatchConfig 从Nacos加载配置并监听配置变化
func (ins *Instance) LoadAndWatchConfig(dataId, group string, onChange func(string)) error {
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

		// 执行自定义回调
		if onChange != nil {
			onChange(data)
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
	// 这里应该解析配置内容并应用到应用程序
	// 根据你的实际需求实现配置解析逻辑
	// 例如：解析JSON/YAML格式的配置并更新全局配置变量

	log.Printf("applying config: %s", content)

	// TODO: 实现具体的配置解析和应用逻辑
	// 示例：
	if err := config.ParseAndApply(content, dataId); err != nil {
		return err
	}

	return nil
}
