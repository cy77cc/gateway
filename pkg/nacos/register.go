package nacos

import (
	"net"
	"os"
	"strconv"

	"github.com/nacos-group/nacos-sdk-go/vo"
)

// Register implements discovery.ServiceRegistry
func (ins *Instance) Register(serviceName string, host string, port int, metadata map[string]string) error {
	// 如果 host 为空，自动获取本机IP
	if host == "" {
		var err error
		host, err = getLocalIP()
		if err != nil {
			return err
		}
	}

	// 如果端口未指定，尝试从环境变量获取
	if port == 0 {
		port = getPortFromEnv()
	}

	if metadata == nil {
		metadata = make(map[string]string)
	}

	// 注册实例到Nacos
	_, err := ins.NamingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          host,
		Port:        uint64(port),
		ServiceName: serviceName,
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata:    metadata,
	})

	return err
}

// Deregister implements discovery.ServiceRegistry
func (ins *Instance) Deregister(serviceName string, host string, port int) error {
	_, err := ins.NamingClient.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          host,
		Port:        uint64(port),
		ServiceName: serviceName,
		Ephemeral:   true,
	})
	return err
}

// getLocalIP 获取本机IP地址
func getLocalIP() (string, error) {
	// 优先通过环境变量获取
	if ip := os.Getenv("SERVER_IP"); ip != "" {
		return ip, nil
	}

	// 自动获取第一个非回环IP
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

// getPortFromEnv 从环境变量获取端口号
func getPortFromEnv() int {
	portStr := os.Getenv("SERVER_PORT")
	if portStr == "" {
		return 8080 // 默认端口
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 8080 // 转换失败使用默认端口
	}

	return port
}
