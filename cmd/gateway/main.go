package main

import (
	"flag"
	"fmt"
	"gateway-demo/config"
	"gateway-demo/internal/discovery"
	"gateway-demo/internal/proxy"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "etc/config.yaml", "config path")

	var gatewayConfigPath string
	flag.StringVar(&gatewayConfigPath, "gateway-config", "etc/gateway-router.json", "gateway config path")

	cfg, err := config.Load(configPath, gatewayConfigPath)

	if err != nil {
		log.Fatal("配置文件为空", err)
	}

	config.CONFIG = cfg

	gin.SetMode(config.CONFIG.Server.Mode)

	r := gin.Default()

	// 初始化 Nacos
	nacosClient := discovery.NewNacosClient()

	// 注册配置文件中的路由
	for _, route := range config.CONFIG.RouteCfg.Routes {
		r.Any(route.PathPrefix+"/*path", proxy.NewRouteProxyHandler(nacosClient, route.Service, route.StripPrefix))
	}

	// 泛路由：/api/:service/*path (作为兜底或开发调试用)
	r.Any("/api/:service/*path", proxy.NewHTTPProxyHandler(nacosClient))

	// 启动
	_ = r.Run(fmt.Sprintf("%s:%d", config.CONFIG.Server.Host, config.CONFIG.Server.Port))
}
