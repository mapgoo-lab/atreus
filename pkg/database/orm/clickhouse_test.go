package orm

import (
	"context"
	"fmt"
	"github.com/mapgoo-lab/atreus/pkg/conf/env"
	"github.com/mapgoo-lab/atreus/pkg/net/trace"
	"github.com/mapgoo-lab/atreus/pkg/net/trace/zipkin"
	"gorm.io/gorm"
	"testing"
	"time"
)

//CREATE TABLE test1
//(
//   `id` UInt64,
//   `info` varchar(100) DEFAULT '',
//   `updated_at` DateTime DEFAULT now(),
//   `created_at` UInt64
//)
//ENGINE = MergeTree
//ORDER BY id

type Test1 struct {
	ID         uint64    `gorm:"column:id;type:UInt64" json:"id"`
	Info       string    `gorm:"column:info;type:varchar(100);;default:''" json:"info"`
	CreatedAt  uint64    `gorm:"column:created_at;type:UInt64;" json:"createdAt"`
	FupdatedAt time.Time `gorm:"column:updated_at;default:now()" json:"updatedAt"`
}

func (a *Test1) TableName() string {
	return "test1"
}

func TestClickHouse(t *testing.T) {
	var dsn = "clickhouse://root:123456@127.0.0.1:19000/test?dial_timeout=10s&read_timeout=10s"
	config := &Config{
		DSN:         dsn,
		Active:      0,
		Idle:        0,
		IdleTimeout: 0,
		GormConfig:  &gorm.Config{},
	}
	db := NewClickhouse(config)

	zConfig := &zipkin.Config{
		Endpoint:      "http://127.0.0.1:9411/api/v2/spans",
		DisableSample: false,
	}
	env.AppID = "gorm-clickhouse"

	zipkin.Init(zConfig)

	session := db.WithContext(trace.NewContext(context.Background(), trace.New("clickhousetest")))

	err := session.Model(&Test1{}).Create(Test1{
		ID:        1,
		Info:      "123",
		CreatedAt: uint64(time.Now().Unix()),
	}).Error

	fmt.Println(err)

	var list []Test1
	err = session.Model(&Test1{}).Find(&list).Error
	fmt.Println(err, list)

	err = session.Model(&Test1{}).Where("id = ?", 1).Update("info", "abcdefg").Error
	fmt.Println(err, list)

	err = session.Model(&Test1{}).Delete("id", 1).Error
	fmt.Println(err)

	time.Sleep(3 * time.Second)
}
