# v1 user

## 项目简介
1. 网关模块：http接口请求转发（仅做转发，不受其他服务变化影响）、jwt鉴权。

## 项目说明
1. 项目接口提供grpc和http两种方式。
2. grpc接口供服务间调用，服务间open api鉴权。
3. http接口供外部（api-gateway）调用。

## TODO
1. 协议转换、黑白名单、负载均衡、限流、服务降级、熔断等。
2. 优化获取http的address信息（使用resolver）
3. jwt鉴权服务使用grpc。
4. 配置平台化。
