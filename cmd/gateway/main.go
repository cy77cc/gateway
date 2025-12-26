package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cy77cc/gateway/config"
	"github.com/cy77cc/gateway/internal/middleware"
	"github.com/cy77cc/gateway/internal/proxy"
	"github.com/cy77cc/gateway/internal/router"
	"github.com/cy77cc/gateway/pkg/loadbalance"
	"github.com/cy77cc/hioshop/common/log"
	"github.com/cy77cc/hioshop/common/register"
	"github.com/cy77cc/hioshop/common/register/types"

	"github.com/cy77cc/hioshop/common/register/nacos"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	ctx := context.Background()
	var configPath string
	flag.StringVar(&configPath, "config", "etc/config.yaml", "config path")

	var gatewayConfigPath string
	flag.StringVar(&gatewayConfigPath, "gateway-router", "etc/gateway-router.json", "gateway config path")

	flag.Parse()

	_ = godotenv.Load(".env")

	// 创建配置管理器
	configManager := config.NewConfigManager()

	// 1. Load Local Config First (Base)
	localConfig, err := config.LoadLocalConfig(configPath)
	if err != nil {
		log.Warnf("Failed to load local config: %v", err)
	} else {
		configManager.SetLocalConfig(localConfig)
	}

	// 加载本地路由配置（如果存在）
	routes, err := config.LoadRoutesFromJSON(gatewayConfigPath)
	if err != nil {
		log.Warnf("Failed to load local routes: %v", err)
	} else {
		remoteConfig := &config.RemoteConfig{
			Routes: routes,
		}
		configManager.SetRemoteConfig(remoteConfig)
	}

	// 2. Load Nacos Config (Overlay)
	cfg := configManager.GetConfig()
	log.SetLevel(log.DEBUG)
	// 2. Load Nacos Config (Overlay)

	nacosCfg := nacos.NewNacosConfig()
	nacosCfg.LoadNacosEnv()

	// Attempt to connect to Nacos
	nacosInstance, err := register.NewRegister(ctx,
		"nacos",
		register.WithEndpoints(nacosCfg.Endpoints...),
		register.WithAuth(nacosCfg.Username, nacosCfg.Password),
		register.WithNamespace(nacosCfg.Namespace),
	)

	if err == nil {
		log.Info("Connected to Nacos")

		item, err := nacosInstance.GetConfig(ctx, "gateway", "DEFAULT_GROUP")
		if err != nil {
			log.Errorf("Failed to load global config from Nacos: %v", err)
		}

		// Load remote configs
		if err := configManager.ParseRemoteConfig(item); err != nil {
			log.Errorf("Failed to load global config from Nacos: %v", err)
		}
		log.Info("Loaded global config from Nacos")

		item, err = nacosInstance.GetConfig(ctx, "gateway-router", "DEFAULT_GROUP")
		if err != nil {
			log.Errorf("Failed to load router config from Nacos: %v", err)
		}

		if err := configManager.ParseRemoteConfig(item); err != nil {
			log.Errorf("Failed to load router config from Nacos: %v", err)
		}

		log.Info("Loaded router config from Nacos")

	} else {
		log.Errorf("Nacos connection failed or not configured: %v. Running in local mode.", err)
	}

	if cfg == nil || (cfg.Server.Port == 0 && cfg.Server.Host == "") {
		// Try to use default if nothing loaded
		log.Warn("Config is empty, using defaults")
		if cfg == nil {
			cfg = &config.MergedConfig{}
		}
		cfg.Server.Port = 8080
		cfg.Server.Mode = "debug"
	}

	gin.SetMode(cfg.Server.Mode)
	r := gin.Default()

	// Setup Proxy
	lb := loadbalance.NewRoundRobin()
	middleware.InitBreakerManager()
	middleware.InitBucketManager()

	proxyHandler := proxy.NewProxyHandler(nacosInstance, lb)

	configManager.RegisterWatcher(proxyHandler)

	// 注册路由时使用当前配置
	currentConfig := configManager.GetConfig()

	newRouter := router.NewRouter()

	// 注册路由
	newRouter.RegisterRoutes(r, currentConfig.Routes, proxyHandler)
	// Seed proxy handler with current config
	proxyHandler.OnConfigChange(currentConfig)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// Start Server
	go func() {
		log.Infof("Starting server on %s", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Service Registration
	if nacosInstance != nil {
		go func() {
			// Give server a moment to start
			time.Sleep(1 * time.Second)
			err := nacosInstance.Register(ctx, &types.ServiceInstance{
				ID:          "gateway",
				ServiceName: "gateway",
				Host:        cfg.Server.Host,
				Port:        cfg.Server.Port,
				Metadata:    map[string]string{"version": "1.0.0"},
				Weight:      1.0,
				GroupName:   "DEFAULT_GROUP",
				ClusterName: "DEFAULT",
			})
			if err != nil {
				log.Errorf("Failed to register service: %v", err)
			} else {
				log.Info("Service registered to Nacos")
			}
		}()
	}

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Warn("Shutting down server...")

	if nacosInstance != nil {
		_ = nacosInstance.Deregister(ctx, "gateway")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
		_ = nacosInstance.Close()
	}()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Info("Server exiting")
}
