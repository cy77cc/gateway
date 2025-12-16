package config

import "time"

type Config struct {
	Server  ServerConfig  `yaml:"server"`
	Nacos   NacosConfig   `yaml:"nacos"`
	Proxy   ProxyConfig   `yaml:"proxy"`
	Routes  []RouteConfig `yaml:"routes"`
	Logging LogConfig     `yaml:"logging"`
}

type ServerConfig struct {
	Name string `yaml:"name"`
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

type NacosConfig struct {
	Addr        string `yaml:"addr"`
	Port        uint64 `yaml:"port"`
	Namespace   string `yaml:"namespace"`
	Group       string `yaml:"group"`
	TimeoutMs   uint64 `yaml:"timeout_ms"`
	AccessKey   string `yaml:"access_key"`
	SecretKey   string `yaml:"secret_key"`
	ContextPath string `yaml:"context_path"`
}

type ProxyConfig struct {
	Timeout   time.Duration `yaml:"timeout_ms"`
	KeepAlive bool          `yaml:"keep_alive"`
}

type RouteConfig struct {
	PathPrefix  string `yaml:"path_prefix"`
	Service     string `yaml:"service"`
	StripPrefix string `yaml:"strip_prefix"`
}

type LogConfig struct {
	Level     string `yaml:"level"`
	AccessLog bool   `yaml:"access_log"`
}
