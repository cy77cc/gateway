package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/cy77cc/gateway/config"
	"github.com/cy77cc/gateway/internal/proxy"
	"github.com/cy77cc/gateway/pkg/discovery"
	"github.com/cy77cc/gateway/pkg/loadbalance"
	"github.com/cy77cc/hioshop/common/log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cy77cc/hioshop/common/nacos"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "etc/config.yaml", "config path")

	var gatewayConfigPath string
	flag.StringVar(&gatewayConfigPath, "gateway-config", "etc/gateway-router.json", "gateway config path")

	flag.Parse()

	_ = godotenv.Load(".env")

	// 1. Load Local Config First (Base)
	if _, err := config.LoadFromFile(configPath, gatewayConfigPath); err != nil {
		log.Errorf("Local config not fully loaded (this is fine if using Nacos only): %v", err)
	}

	// 2. Load Nacos Config (Overlay)
	cfg := config.Get()
	log.SetLevel(log.DEBUG)
	cfg.Nacos.LoadNacosEnv()

	var discoveryService discovery.ServiceDiscovery
	var registryService discovery.ServiceRegistry

	// Attempt to connect to Nacos
	nacosInstance, err := nacos.NewNacosInstance(cfg.Nacos)
	if err == nil {
		log.Info("Connected to Nacos")
		// Load remote configs
		if err := nacosInstance.LoadAndWatchConfig("gateway-global", "DEFAULT_GROUP", cfg.MiddlewareCfg); err != nil {
			log.Errorf("Failed to load global config from Nacos: %v", err)
		}
		if err := nacosInstance.LoadAndWatchConfig("gateway-router", "DEFAULT_GROUP", cfg.RouteCfg); err != nil {
			log.Errorf("Failed to load router config from Nacos: %v", err)
		}

		discoveryService = nacosInstance
		registryService = nacosInstance
	} else {
		log.Errorf("Nacos connection failed or not configured: %v. Running in local mode.", err)
	}

	if cfg == nil || (cfg.Server.Port == 0 && cfg.Server.Host == "") {
		// Try to use default if nothing loaded
		log.Warn("Config is empty, using defaults")
		if cfg == nil {
			cfg = &config.Config{}
		}
		cfg.Server.Port = 8080
		cfg.Server.Mode = "debug"
	}

	gin.SetMode(cfg.Server.Mode)
	r := gin.Default()

	// Setup Proxy
	lb := loadbalance.NewRoundRobin()

	if discoveryService == nil {
		log.Warn("Warning: Service Discovery is not available. Proxying by service name will fail unless you implement a local discovery fallback.")
	}

	proxyHandler := proxy.NewProxyHandler(discoveryService, lb)

	// Register Routes
	for _, route := range cfg.RouteCfg.Routes {
		r.Any(route.PathPrefix+"/*path", proxyHandler.HandleRoute(route.Service, route.StripPrefix))
	}

	// Default/Fallback Route
	r.Any("/api/:service/*path", proxyHandler.HandleGeneric)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// Start Server
	go func() {
		log.Info("Starting server on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Service Registration
	if registryService != nil {
		go func() {
			// Give server a moment to start
			time.Sleep(1 * time.Second)
			err := registryService.Register("gateway", cfg.Server.Host, cfg.Server.Port, nil)
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

	if registryService != nil {
		_ = registryService.Deregister("gateway", cfg.Server.Host, cfg.Server.Port)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Info("Server exiting")
}
