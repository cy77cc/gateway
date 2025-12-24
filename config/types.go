package config

import (
	"errors"
	"time"

	"github.com/cy77cc/hioshop/common/middleware/redis"
)

// LocalConfig 本地配置结构
type LocalConfig struct {
	Server  ServerConfig `yaml:"server"`
	Proxy   ProxyConfig  `yaml:"proxy"`
	Logging LogConfig    `yaml:"logging"`
}

var (
	ErrInvalidRouteConfig = errors.New("invalid route config")
)

// RemoteConfig 远程配置结构
type RemoteConfig struct {
	Routes     []Route          `json:"routes"`
	Middleware MiddlewareConfig `yaml:"middleware"`
}

// MergedConfig 合并后的完整配置
type MergedConfig struct {
	Server     ServerConfig
	Proxy      ProxyConfig
	Logging    LogConfig
	Routes     []Route
	Middleware MiddlewareConfig
}

type ServerConfig struct {
	Name string `yaml:"name"`
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

type ProxyConfig struct {
	Timeout   time.Duration `yaml:"timeout_ms"`
	KeepAlive bool          `yaml:"keep_alive"`
}

type Route struct {
	PathPrefix           string           `yaml:"path_prefix" json:"path_prefix"`
	Service              string           `yaml:"service" json:"service"`
	StripPrefix          string           `yaml:"strip_prefix" json:"strip_prefix"`
	CircuitBreakerConfig *CircuitConfig   `yaml:"circuit" json:"circuit"`
	RateLimitConfig      *RateLimitConfig `yaml:"rate_limit" json:"rate_limit"`
}

type LogConfig struct {
	Level     string `yaml:"level"`
	AccessLog bool   `yaml:"access_log"`
}

type MiddlewareConfig struct {
	Mysql struct {
		Host   string `yaml:"host" json:"host"`
		Port   int    `yaml:"port" json:"port"`
		User   string `yaml:"user" json:"user"`
		Pass   string `yaml:"pass" json:"pass"`
		DBName string `yaml:"db_name" json:"db_name"`
	} `yaml:"mysql"`
	Redis struct {
		Addrs    []string   `yaml:"addrs" json:"addrs"`
		Port     int        `yaml:"port" json:"port"`
		Password string     `yaml:"password" json:"password"`
		Username string     `yaml:"username" json:"username"`
		Type     redis.TYPE `yaml:"type" json:"type"`
	} `yaml:"redis"`
}

type CircuitConfig struct {
	MaxFailures     int     `yaml:"max_failures"`
	MinRequest      int64   `yaml:"min_request"`
	ErrorRate       float64 `yaml:"error_rate"`
	OpenSeconds     int     `yaml:"open_seconds"`
	HalfOpenSuccess int64   `yaml:"half_open_success"`
}

type RateLimitConfig struct {
	Burst    int   `yaml:"burst" json:"burst"`
	QPS      int   `yaml:"qps" json:"qps"`
	LastTime int64 `yaml:"last_time" json:"last_time"`
}
