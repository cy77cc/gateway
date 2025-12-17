package main

import (
	"flag"
	"fmt"
	"gateway/config"
	"gateway/internal/proxy"
	"gateway/pkg/nacos"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/joho/godotenv"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "etc/config.yaml", "config path")

	var gatewayConfigPath string
	flag.StringVar(&gatewayConfigPath, "gateway-config", "etc/gateway-router.json", "gateway config path")

	_ = godotenv.Load(".env")

	config.LoadEnv()

	nacosInstance := nacos.NewNacosInstance(&config.CONFIG.Nacos)

	err := nacosInstance.LoadAndWatchConfig("gateway-global.yaml", "DEFAULT_GROUP", nil)

	// 如果从nacos获取失败，从本地获取
	if err != nil {
		cfg, err := config.Load(configPath, gatewayConfigPath)
		if err != nil {
			log.Fatal("加载配置文件失败")
		}

		config.CONFIG = cfg
	}

	// 获取路由信息
	err = nacosInstance.LoadAndWatchConfig("gateway-router.json", "DEFAULT_GROUP", nil)
	if err != nil {
		log.Fatal("加载配置gateway-global.yaml文件失败")
	}

	gin.SetMode(config.CONFIG.Server.Mode)

	r := gin.Default()

	// 初始化 Nacos

	// 服务注册
	err = nacosInstance.Register("gateway", config.CONFIG.Server.Port)

	if err != nil {
		log.Fatal("注册服务失败", err)
	}

	// 注册配置文件中的路由
	for _, route := range config.CONFIG.RouteCfg.Routes {
		r.Any(route.PathPrefix+"/*path", proxy.NewRouteProxyHandler(nacosInstance, route.Service, route.StripPrefix))
	}

	// 泛路由：/api/:service/*path (作为兜底或开发调试用)
	r.Any("/api/:service/*path", proxy.NewHTTPProxyHandler(nacosInstance))

	// 启动
	_ = r.Run(fmt.Sprintf("%s:%d", config.CONFIG.Server.Host, config.CONFIG.Server.Port))
}
