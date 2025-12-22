```cmd
gateway
├── cmd
│   └── gateway
│       └── main.go
├── internal
│   ├── router        // 路由匹配
│   ├── discovery     // nacos
│   ├── proxy         // http / grpc proxy
│   ├── ratelimit     // 限流
│   ├── circuitbreaker// 熔断
│   ├── retry         // 重试
│   ├── middleware   // gin middleware
│   ├── config       // 动态配置
│   └── metrics      // prometheus
├── pkg
│   ├── slidingwindow
│   ├── tokenbucket
│   └── lb            // round-robin
└── go.mod
```

1. 接收请求
2. 限流 Middleware
3. 鉴权 Middleware
4. 熔断检查
5. 服务发现
6. 负载均衡
7. 请求转发（HTTP / RPC）
8. 统计结果（成功/失败/RT）

需要配置如下环境变量
NACOS_ADDR=55.5.5.5
NACOS_PORT=8848
NACOS_NAMESPACEID=xxx
NACOS_USERNAME=nacos
NACOS_PASSWORD=xxxx
NACOS_CONTEXT_PATH=/nacos
NACOS_AUTH_IDENTITY_KEY=xxx
NACOS_AUTH_IDENTITY_VALUE=xxx
NACOS_AUTH_TOKEN=xxx
