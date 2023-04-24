package orm

import (
	"github.com/mapgoo-lab/atreus/pkg/log"
	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
)

func NewClickhouse(config *Config) *gorm.DB {
	if config.GormConfig.Logger == nil {
		config.GormConfig.Logger = NewLogger()
	}

	db, err := gorm.Open(clickhouse.Open(config.DSN), config.GormConfig)
	if err != nil {
		log.Error("failed to connect clickhouse database error: %s", err.Error())
		return nil
	}

	//连接池
	sqlDb, err := db.DB()
	if err != nil {
		log.Error("failed to get clickhouse db error: %s", err.Error())
		return nil
	}
	sqlDb.SetMaxOpenConns(config.Active)
	sqlDb.SetMaxIdleConns(config.Idle)
	sqlDb.SetConnMaxLifetime(config.IdleTimeout)

	err = db.Use(&ObsPlugin{})
	if err != nil {
		log.Error("Use OrmObsPulgin error: %s", err)
		return nil
	}

	return db
}
