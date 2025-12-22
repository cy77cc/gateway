package config

import (
	"encoding/json"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadLocalConfig 从文件加载本地配置
func LoadLocalConfig(configPath string) (*LocalConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	config := &LocalConfig{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}

// LoadRoutesFromJSON 从JSON文件加载路由配置
func LoadRoutesFromJSON(gatewayConfigPath string) ([]Route, error) {
	data, err := os.ReadFile(gatewayConfigPath)
	if err != nil {
		return nil, err
	}

	var tmp map[string][]Route
	if err := json.Unmarshal(data, &tmp); err != nil {
		return nil, err
	}

	routes, ok := tmp["routes"]
	if !ok {
		return nil, ErrInvalidRouteConfig
	}

	return routes, nil
}

// ParseRemoteRoutes 解析远程路由配置
func ParseRemoteRoutes(content string) ([]Route, error) {
	var routes []Route
	if err := json.Unmarshal([]byte(content), &routes); err != nil {
		return nil, err
	}
	return routes, nil
}

// ParseRemoteMiddleware 解析远程中间件配置
func ParseRemoteMiddleware(content string) (*MiddlewareConfig, error) {
	middleware := &MiddlewareConfig{}
	if err := yaml.Unmarshal([]byte(content), middleware); err != nil {
		return nil, err
	}
	return middleware, nil
}
