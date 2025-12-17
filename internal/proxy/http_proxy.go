package proxy

import (
	"fmt"
	"gateway-demo/pkg/nacos"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

func NewHTTPProxyHandler(nacos *nacos.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		service := c.Param("service")
		path := c.Param("path")

		target, err := nacos.GetAllServiceInstances(service)
		if err != nil {
			c.JSON(502, gin.H{"error": err.Error()})
			return
		}

		// TODO 负载均衡
		targetURL, _ := url.Parse(fmt.Sprintf("http://%s:%s", target[0].Ip, target[0].Port))
		proxy := httputil.NewSingleHostReverseProxy(targetURL)

		// 重写路径
		c.Request.URL.Path = path
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

func NewRouteProxyHandler(nacos *nacos.Instance, serviceName string, stripPrefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		target, err := nacos.GetAllServiceInstances(serviceName)
		if err != nil {
			c.JSON(502, gin.H{"error": err.Error()})
			return
		}

		targetURL, _ := url.Parse(fmt.Sprintf("http://%s:%s", target[0].Ip, target[0].Port))
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
