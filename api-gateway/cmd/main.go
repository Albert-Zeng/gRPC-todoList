package main

import (
	"api-gateway/config"
	"api-gateway/discovery"
	"api-gateway/middleware"
	"api-gateway/pkg/util"
	"api-gateway/proxy"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
	"time"
)

func main() {
	config.InitConfig()
	// etcd注册
	etcdAddress := []string{viper.GetString("etcd.address")}
	etcdRegister := discovery.NewRegister(etcdAddress, logrus.New())
	defer etcdRegister.Stop()
	err := proxy.WatchEtcdServer(etcdRegister)
	if err != nil {
		panic(err)
	}
	//etcdRegister := discovery.NewResolver(etcdAddress, logrus.New())
	//resolver.Register(etcdRegister)
	//defer etcdRegister.Close()

	httpAddress := viper.GetString("server.httpAddress")
	gin.SetMode(gin.DebugMode)
	ginRouter := gin.Default()
	ginRouter.Use(middleware.Cors())
	ginRouter.Use(middleware.JWT())
	//ginRouter.Use(middleware.Cors(), middleware.InitMiddleware(service), middleware.ErrorMiddleware())
	//store := cookie.NewStore([]byte("something-very-secret"))
	//ginRouter.Use(sessions.Sessions("mysession", store))
	ginRouter.Use(func() gin.HandlerFunc {
		return func(c *gin.Context) {
			proxy.RoutesProxy(c)
		}
	}())
	httpServer := &http.Server{
		Addr:           httpAddress,
		Handler:        ginRouter,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	go func() {
		// 优雅关闭
		util.GracefullyShutdown(httpServer)
	}()
	logrus.Info("http server started listen on ", httpAddress)
	err = httpServer.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
