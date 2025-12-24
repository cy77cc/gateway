package svc

import (
	"github.com/cy77cc/gateway/config"
	commonRedis "github.com/cy77cc/hioshop/common/middleware/redis"
	"github.com/cy77cc/hioshop/common/register"
	"github.com/redis/go-redis/v9"
)

type ServiceContext struct {
	Config config.MergedConfig
	// TODO 在common模块创建一个通用的注册中心
	Register *register.Register
	Redis    redis.UniversalClient
}

func NewServiceContext(c config.MergedConfig) *ServiceContext {
	redisComOptions := commonRedis.DefaultCommonOptions()
	redisComOptions.Addrs = c.Middleware.Redis.Addrs
	redisComOptions.Password = c.Middleware.Redis.Password
	redisCfg := commonRedis.Config{
		Type:   c.Middleware.Redis.Type,
		Common: redisComOptions,
	}
	rdb := commonRedis.MustNewRedisClient(&redisCfg)
	return &ServiceContext{
		Config:   c,
		Register: register.NewRegister(),
		Redis:    rdb,
	}
}
