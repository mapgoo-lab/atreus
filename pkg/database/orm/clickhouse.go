package orm

import (
	"github.com/mapgoo-lab/atreus/pkg/log"
	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
)

func NewClickhouse(dsn string, config *gorm.Config) *gorm.DB {
	if config.Logger == nil {
		config.Logger = NewLogger()
	}

	db, err := gorm.Open(clickhouse.Open(dsn), config)
	if err != nil {
		log.Error("failed to connect clickhouse database error: %s", err.Error())
		return nil
	}

	err = db.Use(&ObsPlugin{})
	if err != nil {
		log.Error("Use OrmObsPulgin error: %s", err)
		return nil
	}

	return db
}
