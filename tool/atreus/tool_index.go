package main

import "time"

var toolIndexs = []*Tool{
	&Tool{
		Name:      "atreus",
		Alias:     "atreus",
		BuildTime: time.Date(2019, 6, 21, 0, 0, 0, 0, time.Local),
		Install:   "go get -u github.com/bilibili/atreus/tool/atreus",
		Summary:   "Kratos工具集本体",
		Platform:  []string{"darwin", "linux", "windows"},
		Author:    "atreus",
	},
	&Tool{
		Name:      "protoc",
		Alias:     "atreus-protoc",
		BuildTime: time.Date(2019, 6, 21, 0, 0, 0, 0, time.Local),
		Install:   "go get -u github.com/bilibili/atreus/tool/atreus-protoc",
		Summary:   "快速方便生成pb.go的protoc封装，windows、Linux请先安装protoc工具",
		Platform:  []string{"darwin", "linux", "windows"},
		Author:    "atreus",
	},
	&Tool{
		Name:      "swagger",
		Alias:     "swagger",
		BuildTime: time.Date(2019, 5, 5, 0, 0, 0, 0, time.Local),
		Install:   "go get -u github.com/go-swagger/go-swagger/cmd/swagger",
		Summary:   "swagger api文档",
		Platform:  []string{"darwin", "linux", "windows"},
		Author:    "goswagger.io",
	},
	&Tool{
		Name:      "genbts",
		Alias:     "atreus-gen-bts",
		BuildTime: time.Date(2019, 7, 23, 0, 0, 0, 0, time.Local),
		Install:   "go get -u github.com/bilibili/atreus/tool/atreus-gen-bts",
		Summary:   "缓存回源逻辑代码生成器",
		Platform:  []string{"darwin", "linux", "windows"},
		Author:    "atreus",
	},
	&Tool{
		Name:      "genmc",
		Alias:     "atreus-gen-mc",
		BuildTime: time.Date(2019, 7, 23, 0, 0, 0, 0, time.Local),
		Install:   "go get -u github.com/bilibili/atreus/tool/atreus-gen-mc",
		Summary:   "mc缓存代码生成",
		Platform:  []string{"darwin", "linux", "windows"},
		Author:    "atreus",
	},
}
