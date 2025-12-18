package config

import "time"

type Config struct {
	Server        ServerConfig `yaml:"server"`
	Proxy         ProxyConfig  `yaml:"proxy"`
	Logging       LogConfig    `yaml:"logging"`
	Nacos         *NacosConfig
	RouteCfg      *RouteConfig
	MiddlewareCfg *MiddlewareConfig
}

type ServerConfig struct {
	Name string `yaml:"name"`
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

type NacosConfig struct {
	Endpoint    string `yaml:"endpoint"`
	Port        uint64 `yaml:"port"`
	Namespace   string `yaml:"namespace"`
	Group       string `yaml:"group"`
	TimeoutMs   uint64 `yaml:"timeout_ms"`
	Username    string `yaml:"Username"`
	Password    string `yaml:"Password"`
	ContextPath string `yaml:"context_path"`
	IdentityKey string `yaml:"identity_key"`
	IdentityVal string `yaml:"identity_val"`
	Token       string `yaml:"token"`
}

type ProxyConfig struct {
	Timeout   time.Duration `yaml:"timeout_ms"`
	KeepAlive bool          `yaml:"keep_alive"`
}

type RouteConfig struct {
	Routes []Route `yaml:"routes" json:"routes"`
}

type Route struct {
	PathPrefix  string `yaml:"path_prefix" json:"path_prefix"`
	Service     string `yaml:"service" json:"service"`
	StripPrefix string `yaml:"strip_prefix" json:"strip_prefix"`
}

type LogConfig struct {
	Level     string `yaml:"level"`
	AccessLog bool   `yaml:"access_log"`
}

type MiddlewareConfig struct {
	Mysql struct {
		Host   string `yaml:"host"`
		Port   int    `yaml:"port"`
		User   string `yaml:"user"`
		Pass   string `yaml:"pass"`
		DBName string `yaml:"db_name"`
	} `yaml:"mysql"`
	Redis struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Password string `yaml:"password"`
	} `yaml:"redis"`
}
