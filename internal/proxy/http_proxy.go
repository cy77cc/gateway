package proxy

import (
	"gateway-demo/internal/discovery"
	"net/http/httputil"
	"net/url"

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
