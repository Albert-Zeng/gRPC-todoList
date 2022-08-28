# v1 user

## 项目简介
1. 用户模块：注册、登录、登出、鉴权、查询用户基本信息

## 项目说明
1. 项目接口提供grpc和http两种方式。
2. grpc接口供服务间调用，服务间open api鉴权。
3. http接口供外部（api-gateway）调用，api-gateway的jwt中间件调用user鉴权。

## TODO
1. grpc和http服务间open api鉴权。
2. token生成方式优化，鉴权时返回UserID信息。
3. proto文件内自定义json、form字段名称，参考 github.com/gogo/protobuf/gogoproto/gogo.proto。
