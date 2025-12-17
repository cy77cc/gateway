package nacos

import (
	"fmt"
	"gateway-demo/config"
	"log"

	"github.com/nacos-group/nacos-sdk-go/vo"
)

func (ins *Instance) LoadAndWatchConfig(dataId, group string) {
	content, err := ins.ConfigClient.GetConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
	})

	if err != nil {
		panic(err)
	}

	applyConfig(content)

	err = ins.ConfigClient.ListenConfig(vo.ConfigParam{
		DataId: dataId,
		OnChange: func(namespace, group, dataId, data string) {
			fmt.Println("group:" + group + ", dataId:" + dataId + ", data:" + data)
		},
	})

	if err != nil {
		panic(err)
	}
}
