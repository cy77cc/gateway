package middleware

import (
	"fmt"
	"sync"

	"github.com/cy77cc/gateway/config"
	"github.com/cy77cc/gateway/internal/ratelimit"
	"github.com/gin-gonic/gin"
)

// bucketManager 桶管理器
var bucketManager *BucketManager

// BucketManager 桶管理器结构
type BucketManager struct {
	buckets map[string]*ratelimit.TokenBucket
	mutex   sync.RWMutex
}

// Get 获取指定路由的令牌桶
func (bm *BucketManager) Get(route *config.Route) *ratelimit.TokenBucket {
	if route == nil || route.RateLimitConfig == nil {
		return nil
	}

	key := fmt.Sprintf("%s_%s", route.Service, route.PathPrefix)

	bm.mutex.RLock()
	if bucket, exists := bm.buckets[key]; exists {
		bm.mutex.RUnlock()
		return bucket
	}
	bm.mutex.RUnlock()

	// 创建新的令牌桶
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	// 双重检查
	if bucket, exists := bm.buckets[key]; exists {
		return bucket
	}

	bucket := &ratelimit.TokenBucket{
		capacity: int64(route.RateLimit.Burst),
		tokens:   int64(route.RateLimit.Burst),
		rate:     int64(route.RateLimit.QPS),
		lastTime: 0,
	}

	bm.buckets[key] = bucket
	return bucket
}

// InitBucketManager 初始化桶管理器
func InitBucketManager() {
	bucketManager = &BucketManager{
		buckets: make(map[string]*ratelimit.TokenBucket),
	}
}

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		route := matchRoute(c.Request.URL.Path)
		if route == nil || route.RateLimit == nil {
			c.Next()
			return
		}

		bucket := bucketManager.Get(route)
		if bucket == nil || !bucket.Allow() {
			c.AbortWithStatus(429)
			return
		}

		c.Next()
	}
}

func DistributedRateLimitMiddleware(limiter *ratelimit.RedisRateLimiter) gin.HandlerFunc {
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
