package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"time"
	"user/config"
	"user/discovery"
	"user/internal/cache"
	"user/internal/handler"
	"user/internal/repository"
	"user/internal/service"
	"user/pkg/util"
)

func main() {
	config.InitConfig()
	repository.InitDB()
	cache.InitRedis()
	// etcd 地址
	etcdAddress := []string{viper.GetString("etcd.address")}
	// 服务注册
	etcdRegister := discovery.NewRegister(etcdAddress, logrus.New())
	defer etcdRegister.Stop()

	// grpc
	grpcAddress := viper.GetString("server.grpcAddress")
	server := grpc.NewServer(grpc.UnaryInterceptor(handler.AuthGrpc))
	defer server.Stop()
	// 绑定service
	service.RegisterUserServiceServer(server, handler.NewUserService())
	lis, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		panic(err)
	}
	userGrpcNode := discovery.Server{
		Name:    viper.GetString("server.domain"),
		Protoc:  "grpc",
		Addr:    grpcAddress,
		Version: "v1",
	}
	if _, err := etcdRegister.Register(userGrpcNode, 10); err != nil {
		panic(err)
	}
	logrus.Info("grpc server started listen on ", grpcAddress)
	go func() {
		if err := server.Serve(lis); err != nil {
			panic(err)
		}
	}()

	// http
	conn, err := grpc.Dial(
		grpcAddress,
		grpc.WithInsecure(),
	)
	if err != nil {
		panic(err)
	}
	gwmux := runtime.NewServeMux()
	err = service.RegisterUserServiceHandler(context.Background(), gwmux, conn)
	if err != nil {
		panic(err)
	}
	httpAddress := viper.GetString("server.httpAddress")
	gin.SetMode(gin.DebugMode)
	ginRouter := gin.Default()
	//ginRouter.Use(middleware.Cors(), middleware.InitMiddleware(service), middleware.ErrorMiddleware())
	//store := cookie.NewStore([]byte("something-very-secret"))
	//ginRouter.Use(sessions.Sessions("mysession", store))
	// http auth at api-gateway
	ginRouter.Use(handler.AuthHttp())
	ginRouter.Use(func() gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Status(http.StatusOK)
			gwmux.ServeHTTP(c.Writer, c.Request)
		}
	}())
	httpServer := &http.Server{
		Addr:           httpAddress,
		Handler:        ginRouter,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	userHttpNode := discovery.Server{
		Name:    viper.GetString("server.domain"),
		Protoc:  "http",
		Addr:    httpAddress,
		Version: "v1",
	}
	if _, err := etcdRegister.Register(userHttpNode, 10); err != nil {
		panic(err)
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


