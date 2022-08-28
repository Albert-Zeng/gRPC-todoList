# gRPC-todoList

gin+grpc+gorm+etcd+mysql+redis 的备忘录功能


# 项目主要依赖
- gin
- gorm
- etcd
- grpc
- jwt-go
- logrus
- viper
- protobuf

# 项目结构

## 1. api-gateway 网关部分

```
api-gateway/
├── cmd                   // 启动入口
├── config                // 配置文件
├── discovery             // etcd服务注册、keep-alive、获取服务信息等等
├── internal              // 业务逻辑（网关无业务处理逻辑，仅做转发）
├── logs                  // 放置打印日志模块
├── middleware            // 中间件
├── pkg                   // 各种包
│   ├── e                 // 统一错误状态码
│   ├── res               // 统一response接口返回
│   └── util              // 各种工具、JWT、Logger等等..
├── proxy                 // http请求转发
└── wrappers              // 各个服务之间的熔断降级
```

## 2. user && task 用户与任务模块


```
user/
├── cmd                   // 启动入口
├── config                // 配置文件
├── discovery             // etcd服务注册、keep-alive、获取服务信息等等
├── internal              // 业务逻辑（不对外暴露）
│   ├── dependency        // 外部依赖（第三方）
│   ├── handler           // 视图层
│   ├── cache             // 缓存模块
│   ├── repository        // 持久层
│   └── service           // 服务层
│       └──pb             // 放置生成的pb文件
├── logs                  // 放置打印日志模块
└── pkg                   // 各种包
    ├── e                 // 统一错误状态码
    ├── res               // 统一response接口返回
    └── util              // 各种工具、JWT、Logger等等..
```

# 项目完善
详见各功能模块 README.md。
三个项目可以合并go mod文件，形成"大仓"。

# 项目启动
保证etcd处于运行状态。
- 在各模块下进行

```go
go mod tidy
```

- 在各模块下的cmd目录

```go
go run main.go
```