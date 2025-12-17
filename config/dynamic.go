package config

import "sync/atomic"

var dynamicConfig atomic.Value // *GatewayDynamicConfig

type GatewayDynamicConfig struct {
	Routes []RouteConfig `yaml:"routes"`
}

func InitDynamicConfig(cfg *GatewayDynamicConfig) {
	dynamicConfig.Store(cfg)
}

func GetDynamicConfig() *GatewayDynamicConfig {
	return dynamicConfig.Load().(*GatewayDynamicConfig)
}
