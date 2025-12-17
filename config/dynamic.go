package config

type GatewayDynamicConfig struct {
	Routes []RouteConfig `yaml:"routes"`
}
