package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/cy77cc/gateway/config"
	"github.com/cy77cc/gateway/pkg/loadbalance"
	"github.com/cy77cc/hioshop/common/nacos"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Handler struct {
	discovery    nacos.ServiceDiscovery
	loadBalancer loadbalance.LoadBalancer
}

func NewProxyHandler(d nacos.ServiceDiscovery, lb loadbalance.LoadBalancer) *Handler {
	return &Handler{
		discovery:    d,
		loadBalancer: lb,
	}
}

// HandleGeneric handles /api/:service/*path
func (h *Handler) HandleGeneric(c *gin.Context) {
	service := c.Param("service")
	path := c.Param("path")
	h.proxy(c, service, path, "")
}

// HandleRoute returns a handler for configured routes
func (h *Handler) HandleRoute(serviceName, stripPrefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		h.proxy(c, serviceName, c.Request.URL.Path, stripPrefix)
	}
}

func (h *Handler) proxy(c *gin.Context, serviceName, path, stripPrefix string) {
	// 检查是否是 WebSocket 升级请求
	if h.isWebSocketRequest(c.Request) {
		h.proxyWebSocket(c, serviceName)
		return
	}

	// 原有的 HTTP 反向代理逻辑
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

		targetPath := path
		if stripPrefix != "" && strings.HasPrefix(targetPath, stripPrefix) {
			targetPath = strings.TrimPrefix(targetPath, stripPrefix)
			if !strings.HasPrefix(targetPath, "/") {
				targetPath = "/" + targetPath
			}
		}

		req.URL.Path = targetPath
		req.Host = targetURL.Host
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		if !c.Writer.Written() {
			c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("proxy error: %v", err)})
		}
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

// 判断是否为 WebSocket 升级请求
func (h *Handler) isWebSocketRequest(req *http.Request) bool {
	connectionHeader := strings.ToLower(req.Header.Get("Connection"))
	upgradeHeader := strings.ToLower(req.Header.Get("Upgrade"))
	return strings.Contains(connectionHeader, "upgrade") && upgradeHeader == "websocket"
}

// WebSocket 代理实现
func (h *Handler) proxyWebSocket(c *gin.Context, serviceName string) {
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

	// 构建目标 URL
	targetURL := url.URL{
		Scheme: "ws",
		Host:   fmt.Sprintf("%s:%d", instance.Host, instance.Port),
		Path:   c.Request.URL.Path,
	}

	// 创建 WebSocket 连接
	dialer := websocket.DefaultDialer
	targetConn, resp, err := dialer.Dial(targetURL.String(), c.Request.Header)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("failed to connect to backend: %v", err)})
		if resp != nil {
			resp.Body.Close()
		}
		c.Abort()
		return
	}
	defer targetConn.Close()

	// 升级客户端连接
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许跨域
		},
	}
	clientConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("failed to upgrade client connection: %v", err)})
		c.Abort()
		return
	}
	defer clientConn.Close()

	// 在两个连接之间转发消息
	errChan := make(chan error, 2)

	// 从客户端转发到后端服务
	go func() {
		defer close(errChan)
		for {
			messageType, message, err := clientConn.ReadMessage()
			if err != nil {
				errChan <- err
				return
			}

			if err := targetConn.WriteMessage(messageType, message); err != nil {
				errChan <- err
				return
			}
		}
	}()

	// 从后端服务转发到客户端
	go func() {
		for {
			messageType, message, err := targetConn.ReadMessage()
			if err != nil {
				errChan <- err
				return
			}

			if err := clientConn.WriteMessage(messageType, message); err != nil {
				errChan <- err
				return
			}
		}
	}()

	// 等待任一方向出现错误
	<-errChan
}

func (h *Handler) OnConfigChange(config *config.MergedConfig) {

}
