package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		route := matchRoute(c.Request.URL.Path)
		if route == nil || route.RateLimit == nil {
			c.Next()
			return
		}

		bucket := bucketManager.Get(route)
		if !bucket.Allow() {
			c.AbortWithStatus(429)
			return
		}

		c.Next()
	}
}

func DistributedRateLimitMiddleware(limiter *RedisRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		route := matchRoute(c.Request.URL.Path)
		if route == nil || route.RateLimit == nil {
			c.Next()
			return
		}

		key := fmt.Sprintf(
			"rl:%s:%s",
			route.Service,
			route.PathPrefix,
		)

		ok, err := limiter.Allow(
			c.Request.Context(),
			key,
			route.RateLimit.Burst,
			route.RateLimit.QPS,
		)

		if err != nil {
			// Redis 异常策略：放行 or 拒绝？
			c.Next() // 推荐：放行
			return
		}

		if !ok {
			c.AbortWithStatus(429)
			return
		}

		c.Next()
	}
}
