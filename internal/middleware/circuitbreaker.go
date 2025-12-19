package middleware

import "github.com/gin-gonic/gin"

func CircuitBreakerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		route := matchRoute(c.Request.URL.Path)
		cb := breakerManager.Get(route)

		if !cb.Allow() {
			c.AbortWithStatus(503)
			return
		}

		c.Next()

		success := c.Writer.Status() < 500
		cb.OnResult(success, route.CircuitBreaker)
	}
}
