package ratelimit

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

type TokenBucket struct {
	capacity int64
	tokens   int64
	rate     int64 // tokens per second
	lastTime int64
}

func NewTokenBucket(capacity, rate int64) *TokenBucket {
	return &TokenBucket{
		capacity: capacity,
		tokens:   capacity,
		rate:     rate,
		lastTime: time.Now().UnixNano(),
	}
}

// Allow 检查是否允许请求通过令牌桶限流器
// 该方法会根据时间流逝补充令牌，并尝试消费一个令牌
// 如果当前有可用令牌，则消费一个令牌并返回true；否则返回false
// 返回值: bool - true表示请求被允许，false表示请求被限流
func (b *TokenBucket) Allow() bool {
	now := time.Now().UnixNano()
	last := atomic.LoadInt64(&b.lastTime)

	// 计算时间间隔并补充令牌
	elapsed := (now - last) / int64(time.Second)
	if elapsed > 0 {
		newTokens := elapsed * b.rate
		cur := atomic.LoadInt64(&b.tokens)
		if cur < b.capacity {
			atomic.StoreInt64(&b.tokens, min(b.capacity, cur+newTokens))
		}
		atomic.CompareAndSwapInt64(&b.lastTime, last, now)
	}

	// 原子性地检查并消费令牌
	for {
		cur := atomic.LoadInt64(&b.tokens)
		if cur <= 0 {
			return false
		}
		if atomic.CompareAndSwapInt64(&b.tokens, cur, cur-1) {
			return true
		}
	}
}

type RedisRateLimiter struct {
	client *redis.Client
	script *redis.Script
}

// NewRedisRateLimiter 创建一个新的Redis限流器实例
// 该限流器使用令牌桶算法实现，通过Redis的Lua脚本保证原子性操作
//
// 参数:
//
//	rdb: Redis客户端实例，用于执行限流相关的Redis操作
//
// 返回值:
//
//	*RedisRateLimiter: Redis限流器实例，包含Redis客户端和预编译的Lua脚本
func NewRedisRateLimiter(rdb *redis.Client) *RedisRateLimiter {

	// Lua脚本实现令牌桶算法的核心逻辑
	// 脚本参数说明：
	// KEYS[1]: 限流桶的Redis键名
	// ARGV[1]: 桶的容量（最大令牌数）
	// ARGV[2]: 令牌生成速率（每秒生成的令牌数）
	// ARGV[3]: 当前时间戳（秒）
	//
	// 脚本功能：
	// 1. 获取当前桶中的令牌数和上次更新时间
	// 2. 根据时间差和速率计算应补充的令牌数
	// 3. 判断是否可以消费一个令牌（进行限流控制）
	// 4. 更新桶状态并设置过期时间
	luaTokenBucket := `
-- KEYS[1] = bucket key
-- ARGV[1] = capacity
-- ARGV[2] = rate (tokens per second)
-- ARGV[3] = now (unix timestamp, seconds)

local bucket = redis.call("HMGET", KEYS[1], "tokens", "ts")

local tokens = tonumber(bucket[1])
local ts = tonumber(bucket[2])

if tokens == nil then
    tokens = tonumber(ARGV[1])
    ts = tonumber(ARGV[3])
end

local delta = math.max(0, tonumber(ARGV[3]) - ts)
local filled = math.min(tonumber(ARGV[1]), tokens + delta * tonumber(ARGV[2]))

if filled < 1 then
    redis.call("HMSET", KEYS[1], "tokens", filled, "ts", ARGV[3])
    redis.call("EXPIRE", KEYS[1], 2)
    return 0
end

redis.call("HMSET", KEYS[1], "tokens", filled - 1, "ts", ARGV[3])
redis.call("EXPIRE", KEYS[1], 2)
return 1

`

	return &RedisRateLimiter{
		client: rdb,
		script: redis.NewScript(luaTokenBucket),
	}
}

func (r *RedisRateLimiter) Allow(
	ctx context.Context,
	key string,
	capacity int,
	rate int,
) (bool, error) {

	now := time.Now().Unix()

	res, err := r.script.Run(
		ctx,
		r.client,
		[]string{key},
		capacity,
		rate,
		now,
	).Int()

	if err != nil {
		return false, err
	}
	return res == 1, nil
}
