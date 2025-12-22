package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/cy77cc/hioshop/common/nacos"
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

	err = instance.Register("usercenter", "118.193.38.89", 443, map[string]string{"instance": "1"})
	err = instance.Register("usercenter", "115.190.245.134", 8080, map[string]string{"instance": "2"})
	err = instance.Register("usercenter", "115.190.245.134", 8848, map[string]string{"instance": "3"})
	if err != nil {
		panic(err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

}
