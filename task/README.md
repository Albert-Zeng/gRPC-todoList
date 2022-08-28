# v1 user

## 项目简介
1. 任务模块：任务创建、修改、删除、查询（grpc调用查询用户基本信息）。

## 项目说明
1. 项目接口提供grpc和http两种方式。
2. grpc接口供服务间调用，服务间open api鉴权。
3. http接口供外部（api-gateway）调用。

## TODO
1. grpc和http服务间open api鉴权。
2. 查询usr模块用户基本信息时，优化获取grpc的address信息（使用resolver）
3. proto文件内自定义json、form字段名称，参考 github.com/gogo/protobuf/gogoproto/gogo.proto。
4. 鉴权应该要有UserID返回，获取备忘录时UserID强校验（原则上只能获取自己的列表）。
