package orm

import (
	"github.com/mapgoo-lab/atreus/pkg/log"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

func NewMSSQL(config *Config) *gorm.DB {
	if config.GormConfig == nil {
		config.GormConfig = &gorm.Config{}
	}
	if config.GormConfig.Logger == nil {
		config.GormConfig.Logger = NewLogger()
	}

	db, err := gorm.Open(sqlserver.Open(config.DSN), config.GormConfig)
	if err != nil {
		log.Error("failed to connect mysql database")
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
