package middleware

import (
	"fmt"
	"sync"

	"github.com/cy77cc/gateway/config"
	"github.com/cy77cc/gateway/internal/ratelimit"
	"github.com/gin-gonic/gin"
)

var bucketManager *BucketManager

// BucketManager 桶管理器结构，根据路由创建对应的令牌桶
type BucketManager struct {
	buckets map[string]*ratelimit.TokenBucket
	mutex   sync.RWMutex
}

// Get 根据路由配置获取对应的令牌桶，如果不存在则创建新的令牌桶
// 参数:
//
//	route: 路由配置信息，包含服务名、路径前缀和限流配置
//
// 返回值:
//
//	*ratelimit.TokenBucket: 对应的令牌桶实例，如果路由配置无效则返回nil
func (bm *BucketManager) Get(route *config.Route) *ratelimit.TokenBucket {
	// 如果路由为空或这个路由没有限流配置，则返回nil
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

	// 如果还没有创建过这个路由的令牌桶，则创建
	// 创建新的令牌桶
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	// 双重检查
	if bucket, exists := bm.buckets[key]; exists {
		return bucket
	}

	bucket := ratelimit.NewTokenBucket(int64(route.RateLimitConfig.Burst), int64(route.RateLimitConfig.QPS))

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
		if route == nil || route.RateLimitConfig == nil {
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
		if route == nil || route.RateLimitConfig == nil {
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
			route.RateLimitConfig.Burst,
			route.RateLimitConfig.QPS,
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
