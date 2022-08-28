package proxy

import (
	"api-gateway/discovery"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

var (
	// serverMap version/name/protoc(http/grpc):*Server
	_serverMap = map[string]*discovery.Server{}
)

func WatchEtcdServer(etcdRegister *discovery.Register) error {
	// 监听全部 "/"
	return etcdRegister.WatchEtcdServerInfo(_serverMap, "/")
}

func GetServer(k string) *discovery.Server {
	s := _serverMap
	fmt.Println(s)
	return _serverMap[k]
}

func RoutesProxy(c *gin.Context) {
	split := strings.Split(c.Request.URL.Path, "/")
	// 不存在直接404
	k := c.Request.URL.Path
	if len(split) >= 4 {
		k = fmt.Sprintf("/%s/%s", strings.Join(split[2:4], "/"), "http")
	}
	server := GetServer(k)
	if server == nil {
		c.Status(http.StatusNotFound)
		c.Abort()
		return
	}
	// todo: 负载均衡策略，需要修改(_serverMap)的数据结构
	relay(c, &TargetHost{
		Host: server.Addr,
	})
}