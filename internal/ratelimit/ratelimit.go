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

func (b *TokenBucket) Allow() bool {
	now := time.Now().UnixNano()
	last := atomic.LoadInt64(&b.lastTime)

	elapsed := (now - last) / int64(time.Second)
	if elapsed > 0 {
		newTokens := elapsed * b.rate
		cur := atomic.LoadInt64(&b.tokens)
		if cur < b.capacity {
			atomic.StoreInt64(&b.tokens, min(b.capacity, cur+newTokens))
		}
		atomic.CompareAndSwapInt64(&b.lastTime, last, now)
	}

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

func NewRedisRateLimiter(rdb *redis.Client) *RedisRateLimiter {

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
