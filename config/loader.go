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
func LoadNacosEnv() NacosConfig {
	nacosCfg := NacosConfig{}
	nacosCfg.Endpoint = os.Getenv("NACOS_ADDR")
	if port := os.Getenv("NACOS_PORT"); port != "" {
		nacosCfg.Port, _ = strconv.ParseUint(port, 10, 64)
	} else {
		nacosCfg.Port = 8848
	}

	nacosCfg.Namespace = os.Getenv("NACOS_NAMESPACEID")
	nacosCfg.ContextPath = os.Getenv("NACOS_CONTEXT_PATH")
	if nacosCfg.ContextPath == "" {
		nacosCfg.ContextPath = "/nacos"
	}
	nacosCfg.Username = os.Getenv("NACOS_USERNAME")
	nacosCfg.Password = os.Getenv("NACOS_PASSWORD")
	nacosCfg.IdentityKey = os.Getenv("NACOS_AUTH_IDENTITY_KEY")
	nacosCfg.IdentityVal = os.Getenv("NACOS_AUTH_IDENTITY_VALUE")
	nacosCfg.Token = os.Getenv("NACOS_AUTH_TOKEN")

	lock.Lock()
	cfg.Nacos = nacosCfg
	lock.Unlock()

	return nacosCfg
}

// UpdateRouteConfig updates the route configuration
func UpdateRouteConfig(content string) error {
	lock.Lock()
	defer lock.Unlock()
	return json.Unmarshal([]byte(content), &cfg.RouteCfg)
}

// UpdateServerConfig updates the server configuration
func UpdateServerConfig(content string) error {
	lock.Lock()
	defer lock.Unlock()
	return yaml.Unmarshal([]byte(content), &cfg.Server)
}
