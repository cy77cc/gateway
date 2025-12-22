package middleware

import (
	"github.com/cy77cc/gateway/internal/circuitbreaker"
	"github.com/gin-gonic/gin"
)

// breakerManager 熔断器管理器
var breakerManager *CircuitBreakerManager

// CircuitBreakerManager 熔断器管理器结构
type CircuitBreakerManager struct {
	breakers map[string]*circuitbreaker.CircuitBreaker
}

// Get 获取指定路由的熔断器
func (cm *CircuitBreakerManager) Get(routeKey string) *circuitbreaker.CircuitBreaker {
	if cb, exists := cm.breakers[routeKey]; exists {
		return cb
	}

	// 如果不存在，创建新的熔断器（使用10个时间窗口桶）
	cb := circuitbreaker.NewCircuitBreaker(10)
	cm.breakers[routeKey] = cb
	return cb
}

// InitBreakerManager 初始化熔断器管理器
func InitBreakerManager() {
	breakerManager = &CircuitBreakerManager{
		breakers: make(map[string]*circuitbreaker.CircuitBreaker),
	}
}

func CircuitBreakerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 匹配路由获取路由配置
		route := matchRoute(c.Request.URL.Path)

		// 获取对应路由的熔断器
		cb := breakerManager.Get(route.PathPrefix) // 假设route有唯一的Key标识

		// 检查是否允许请求通过
		if !cb.Allow() {
			// 熔断器开启，返回503服务不可用
			c.AbortWithStatus(503)
			return
		}

		// 继续处理请求
		c.Next()

		// 根据响应状态判断请求是否成功
		success := c.Writer.Status() < 500

		// 更新熔断器统计信息
		// 假设route配置中有CircuitBreaker相关配置
		cb.OnResult(success, route.CircuitBreakerConfig)
	}
}
