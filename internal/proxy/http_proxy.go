package proxy

import (
	"fmt"
	"gateway/pkg/discovery"
	"gateway/pkg/loadbalance"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

type ProxyHandler struct {
	discovery    discovery.ServiceDiscovery
	loadBalancer loadbalance.LoadBalancer
}

func NewProxyHandler(d discovery.ServiceDiscovery, lb loadbalance.LoadBalancer) *ProxyHandler {
	return &ProxyHandler{
		discovery:    d,
		loadBalancer: lb,
	}
}

// HandleGeneric handles /api/:service/*path
func (h *ProxyHandler) HandleGeneric(c *gin.Context) {
	service := c.Param("service")
	path := c.Param("path")
	// For generic proxy, we might want to use the path directly or stripping /api/:service
	// Based on original code: c.Request.URL.Path = path
	// So we pass the captured path.

	// Reconstruct the path to be forwarded
	// If path is "/foo", we want to forward "/foo"
	h.proxy(c, service, path, "")
}

// HandleRoute returns a handler for configured routes
func (h *ProxyHandler) HandleRoute(serviceName, stripPrefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		h.proxy(c, serviceName, c.Request.URL.Path, stripPrefix)
	}
}

func (h *ProxyHandler) proxy(c *gin.Context, serviceName, path, stripPrefix string) {
	instances, err := h.discovery.GetService(serviceName)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("service discovery error: %v", err)})
		c.Abort()
		return
	}

	instance, err := h.loadBalancer.Select(instances)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "no available instances"})
		c.Abort()
		return
	}

	targetURL, _ := url.Parse(fmt.Sprintf("http://%s:%d", instance.Host, instance.Port))
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// Handle StripPrefix logic
		// If stripPrefix is set, we need to operate on the original request path
		targetPath := path
		if stripPrefix != "" && strings.HasPrefix(targetPath, stripPrefix) {
			targetPath = strings.TrimPrefix(targetPath, stripPrefix)
			if !strings.HasPrefix(targetPath, "/") {
				targetPath = "/" + targetPath
			}
		}

		// For generic proxy where path is explicitly passed (and might be different from c.Request.URL.Path)
		// we should respect the passed path argument if it differs, but here 'path' argument IS the intended target path
		// unless stripPrefix modifies it.

		req.URL.Path = targetPath

		// Set Host header to target host to avoid issues with some servers checking Host header
		req.Host = targetURL.Host
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		// Only write error if headers haven't been written
		if !c.Writer.Written() {
			c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("proxy error: %v", err)})
		}
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}
