package config

import (
	"encoding/json"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

var CONFIG *Config

func Load(path, gatewayConfigPath string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	gatewayRouter, err := os.ReadFile(gatewayConfigPath)

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(gatewayRouter, &cfg.RouteCfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func ParseAndApply(content string, dataId string) error {
	switch dataId {
	case "gateway-router.json":
		err := json.Unmarshal([]byte(content), &CONFIG.RouteCfg)
		return err
	case "gateway-global.yaml":
		err := yaml.Unmarshal([]byte(content), &CONFIG.Server)
		return err
	default:
		return nil
	}
}

func LoadEnv() {
	nacosCfg := NacosConfig{}
	nacosCfg.Endpoint = os.Getenv("NACOS_ADDR")
	nacosCfg.Port, _ = strconv.ParseUint(os.Getenv("NACOS_PORT"), 10, 64)
	nacosCfg.Namespace = os.Getenv("NACOS_NAMESPACEID")
	nacosCfg.ContextPath = os.Getenv("NACOS_CONTEXT_PATH")
	nacosCfg.Username = os.Getenv("NACOS_USERNAME")
	nacosCfg.Password = os.Getenv("NACOS_PASSWORD")
	CONFIG.Nacos = nacosCfg
}
