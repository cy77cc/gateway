package config

import (
	"encoding/json"
	"os"

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
