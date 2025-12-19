package config

import (
	"encoding/json"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

import "sync/atomic"

var globalCfg atomic.Value // *Config

func init() {
	cfg := &Config{
		MiddlewareCfg: &MiddlewareConfig{},
		RouteCfg:      &RouteConfig{},
		Nacos:         &NacosConfig{},
	}
	globalCfg.Store(cfg)
}

// Get returns the current configuration safely
func Get() *Config {
	return globalCfg.Load().(*Config)
}

// LoadFromFile loads configuration from local files
func LoadFromFile(configPath, gatewayConfigPath string) (*Config, error) {
	newCfg := &Config{}

	// main config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, newCfg); err != nil {
		return nil, err
	}

	// routes
	gatewayRouter, err := os.ReadFile(gatewayConfigPath)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(gatewayRouter, &newCfg.RouteCfg); err != nil {
		return nil, err
	}

	// 原子替换
	globalCfg.Store(newCfg)
	return newCfg, nil
}

// LoadNacosEnv loads Nacos config from environment variables
func (c *NacosConfig) LoadNacosEnv() {
	c.Endpoint = os.Getenv("NACOS_ADDR")

	if port := os.Getenv("NACOS_PORT"); port != "" {
		c.Port, _ = strconv.ParseUint(port, 10, 64)
	} else {
		c.Port = 8848
	}

	c.Namespace = os.Getenv("NACOS_NAMESPACEID")
	c.ContextPath = os.Getenv("NACOS_CONTEXT_PATH")
	if c.ContextPath == "" {
		c.ContextPath = "/nacos"
	}
	c.Username = os.Getenv("NACOS_USERNAME")
	c.Password = os.Getenv("NACOS_PASSWORD")
	c.IdentityKey = os.Getenv("NACOS_AUTH_IDENTITY_KEY")
	c.IdentityVal = os.Getenv("NACOS_AUTH_IDENTITY_VALUE")
	c.Token = os.Getenv("NACOS_AUTH_TOKEN")
}

// ApplyConfig updates the route configuration
func (c *RouteConfig) ApplyConfig(content string) error {
	old := Get()

	newCfg := *old // 浅拷贝 Config
	newRoute := &RouteConfig{}

	// 2. 解析新路由配置
	if err := json.Unmarshal([]byte(content), newRoute); err != nil {
		return err
	}

	// 3. 替换 RouteCfg
	newCfg.RouteCfg = newRoute

	// 4. 原子替换
	globalCfg.Store(&newCfg)
	return nil
}

// ApplyConfig updates the server configuration
func (c *MiddlewareConfig) ApplyConfig(content string) error {
	old := Get()

	newCfg := *old
	newMiddleware := &MiddlewareConfig{}

	if err := yaml.Unmarshal([]byte(content), newMiddleware); err != nil {
		return err
	}

	newCfg.MiddlewareCfg = newMiddleware
	globalCfg.Store(&newCfg)
	return nil
}
