# Atreus

Atreus是B站开源的Kratos微服务框架修改来的，包含一些BUG修改和自有特性，为了快速修改，特别另外拉一个分支；  

> 名字来源于:B站取得名字叫Kratos（奎爷），本项目是Kratos生的仔，就叫他儿子的名字Atreus（阿特柔斯）

一下内容基本都来自Kratos项目，基本无改动

## Goals

我们致力于提供完整的微服务研发体验，整合相关框架及工具后，微服务治理相关部分可对整体业务开发周期无感，从而更加聚焦于业务交付。对每位开发者而言，整套Kratos框架也是不错的学习仓库，可以了解和参考到[bilibili](https://www.bilibili.com)在微服务方面的技术积累和经验。

## Features
* HTTP Blademaster：核心基于[gin](https://github.com/gin-gonic/gin)进行模块化设计，简单易用、核心足够轻量；
* GRPC Warden：基于官方gRPC开发，集成[discovery](https://github.com/bilibili/discovery)服务发现，并融合P2C负载均衡；
* Cache：优雅的接口化设计，非常方便的缓存序列化，推荐结合代理模式[overlord](https://github.com/bilibili/overlord)；
* Database：集成MySQL/HBase/TiDB，添加熔断保护和统计支持，可快速发现数据层压力；
* Config：方便易用的[paladin sdk](doc/wiki-cn/config.md)，可配合远程配置中心，实现配置版本管理和更新；
* Log：类似[zap](https://github.com/uber-go/zap)的field实现高性能日志库，并结合log-agent实现远程日志管理；
* Trace：基于opentracing，集成了全链路trace支持（gRPC/HTTP/MySQL/Redis/Memcached）；
* Kratos Tool：工具链，可快速生成标准项目，或者通过Protobuf生成代码，非常便捷使用gRPC、HTTP、swagger文档；

## Quick start

### Requirments

Go version>=1.12 and GO111MODULE=on


### Installation
```shell
go get -u github.com/mapgoo-lab/atreus/tool/atreus
cd $GOPATH/src
atreus new atreus-demo
```

通过 `atreus new` 会快速生成基于kratos库的脚手架代码

### Build & Run

```shell
cd atreus-demo/cmd
go build
./cmd -conf ../configs
```

打开浏览器访问：[http://localhost:8000/atreus-demo/start](http://localhost:8000/atreus-demo/start)，你会看到输出了`Golang 大法好 ！！！`

[快速开始](doc/wiki-cn/quickstart.md)  [kratos工具](doc/wiki-cn/atreus-tool.md)

## Documentation

[简体中文](doc/wiki-cn/summary.md)

## License
Atreus is under the MIT license. See the [LICENSE](./LICENSE) file for details.

-------------

*Please report bugs, concerns, suggestions by issues, or join QQ-group 716486124 to discuss problems around source code.*
