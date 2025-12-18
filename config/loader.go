package config

import (
	"encoding/json"
	"os"
	"strconv"
	"sync"

	"gopkg.in/yaml.v3"
)

var (
	cfg  *Config
	lock sync.RWMutex
)

func init() {
	cfg = &Config{}
	cfg.MiddlewareCfg = new(MiddlewareConfig)
	cfg.RouteCfg = new(RouteConfig)
	cfg.Nacos = new(NacosConfig)
}

// Get returns the current configuration safely
func Get() *Config {
	lock.RLock()
	defer lock.RUnlock()
	return cfg
}

// LoadFromFile loads configuration from local files
func LoadFromFile(configPath, gatewayConfigPath string) (*Config, error) {
	newCfg := &Config{}

	// Load main config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, newCfg); err != nil {
		return nil, err
	}

	// Load gateway routes
	gatewayRouter, err := os.ReadFile(gatewayConfigPath)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(gatewayRouter, &newCfg.RouteCfg); err != nil {
		return nil, err
	}

	lock.Lock()
	cfg = newCfg
	lock.Unlock()

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
	lock.Lock()
	defer lock.Unlock()
	return json.Unmarshal([]byte(content), &cfg.RouteCfg)
}

// ApplyConfig updates the server configuration
func (c *MiddlewareConfig) ApplyConfig(content string) error {
	lock.Lock()
	defer lock.Unlock()
	return yaml.Unmarshal([]byte(content), &cfg.Server)
}
