package proxy

import (
	"gateway-demo/internal/nacos/discovery"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

func NewHTTPProxyHandler(nacos *discovery.NacosClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		service := c.Param("service")
		path := c.Param("path")

		target, err := nacos.Resolve(service)
		if err != nil {
			c.JSON(502, gin.H{"error": err.Error()})
			return
		}

		targetURL, _ := url.Parse(target)
		proxy := httputil.NewSingleHostReverseProxy(targetURL)

		// 重写路径
		c.Request.URL.Path = path
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

func NewRouteProxyHandler(nacos *discovery.NacosClient, serviceName string, stripPrefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		target, err := nacos.Resolve(serviceName)
		if err != nil {
			c.JSON(502, gin.H{"error": err.Error()})
			return
		}

		targetURL, _ := url.Parse(target)
		proxy := httputil.NewSingleHostReverseProxy(targetURL)

		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)
			if stripPrefix != "" && strings.HasPrefix(req.URL.Path, stripPrefix) {
				req.URL.Path = strings.TrimPrefix(req.URL.Path, stripPrefix)
				if !strings.HasPrefix(req.URL.Path, "/") {
					req.URL.Path = "/" + req.URL.Path
				}
			}
		}

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
