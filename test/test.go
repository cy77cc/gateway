package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/cy77cc/hioshop/common/register/nacos"
	"github.com/cy77cc/hioshop/common/register/types"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load(".env")
	nacosConfig := nacos.NewNacosConfig()
	nacosConfig.LoadNacosEnv()
	instance, err := nacos.NewNacosInstance(nacosConfig)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	svc := &types.ServiceInstance{
		ServiceName: "usercenter",
		Host:          "118.193.38.89",
		Port:        443,
		Metadata:    map[string]string{"instance": "1"},
	}
	err = instance.Register(ctx, svc)
	if err != nil {
		panic(err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

}
