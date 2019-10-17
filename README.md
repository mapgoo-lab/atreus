![atreus](doc/img/atreus3.png)

[![Language](https://img.shields.io/badge/Language-Go-blue.svg)](https://golang.org/)
[![Build Status](https://travis-ci.org/bilibili/atreus.svg?branch=master)](https://travis-ci.org/bilibili/atreus)
[![GoDoc](https://godoc.org/github.com/mapgoo-lab/atreus?status.svg)](https://godoc.org/github.com/mapgoo-lab/atreus)
[![Go Report Card](https://goreportcard.com/badge/github.com/mapgoo-lab/atreus)](https://goreportcard.com/report/github.com/mapgoo-lab/atreus)

# Atreus

Atreus是[bilibili](https://www.bilibili.com)开源的一套Go微服务框架，包含大量微服务相关框架及工具。  

> 名字来源于:《战神》游戏以希腊神话为背景，讲述由凡人成为战神的奎托斯（Atreus）成为战神并展开弑神屠杀的冒险历程。

## Goals

我们致力于提供完整的微服务研发体验，整合相关框架及工具后，微服务治理相关部分可对整体业务开发周期无感，从而更加聚焦于业务交付。对每位开发者而言，整套Atreus框架也是不错的学习仓库，可以了解和参考到[bilibili](https://www.bilibili.com)在微服务方面的技术积累和经验。

## Features
* HTTP Blademaster：核心基于[gin](https://github.com/gin-gonic/gin)进行模块化设计，简单易用、核心足够轻量；
* GRPC Warden：基于官方gRPC开发，集成[discovery](https://github.com/bilibili/discovery)服务发现，并融合P2C负载均衡；
* Cache：优雅的接口化设计，非常方便的缓存序列化，推荐结合代理模式[overlord](https://github.com/bilibili/overlord)；
* Database：集成MySQL/HBase/TiDB，添加熔断保护和统计支持，可快速发现数据层压力；
* Config：方便易用的[paladin sdk](doc/wiki-cn/config.md)，可配合远程配置中心，实现配置版本管理和更新；
* Log：类似[zap](https://github.com/uber-go/zap)的field实现高性能日志库，并结合log-agent实现远程日志管理；
* Trace：基于opentracing，集成了全链路trace支持（gRPC/HTTP/MySQL/Redis/Memcached）；
* Atreus Tool：工具链，可快速生成标准项目，或者通过Protobuf生成代码，非常便捷使用gRPC、HTTP、swagger文档；

## Quick start

### Requirments

Go version>=1.12 and GO111MODULE=on

### Installation
```shell
go get -u github.com/mapgoo-lab/atreus/tool/atreus
cd $GOPATH/src
atreus new atreus-demo
```

通过 `atreus new` 会快速生成基于atreus库的脚手架代码，如生成 [atreus-demo](https://github.com/mapgoo-lab/atreus-demo) 

### Build & Run

```shell
cd atreus-demo/cmd
go build
./cmd -conf ../configs
```

打开浏览器访问：[http://localhost:8000/atreus-demo/start](http://localhost:8000/atreus-demo/start)，你会看到输出了`Golang 大法好 ！！！`

[快速开始](doc/wiki-cn/quickstart.md)  [atreus工具](doc/wiki-cn/atreus-tool.md)

## Documentation

> [简体中文](doc/wiki-cn/summary.md)  
> [FAQ](doc/wiki-cn/FAQ.md)  

## License
Atreus is under the MIT license. See the [LICENSE](./LICENSE) file for details.

-------------

*Please report bugs, concerns, suggestions by issues, or join QQ-group 716486124 to discuss problems around source code.*
