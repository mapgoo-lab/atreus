package orm

import (
	"gorm.io/gorm"
	"time"
)

// Config database config.
type Config struct {
	DSN         string        // write data source name.
	Active      int           // 开启最大连接数
	Idle        int           // 最大闲置连接数
	IdleTimeout time.Duration // 连接空闲时间
	GormConfig  *gorm.Config
	ReadDSN     []string //读库
}
