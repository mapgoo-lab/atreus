package orm

import (
	"github.com/mapgoo-lab/atreus/pkg/log"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

func NewMSSQL(dsn string, config *gorm.Config) *gorm.DB {
	db, err := gorm.Open(sqlserver.Open(dsn), config)
	if err != nil {
		log.Error("failed to connect mysql database")
		return nil
	}

	err = db.Use(&ObsPlugin{})
	if err != nil {
		log.Error("Use OrmObsPulgin error: %s", err)
		return nil
	}

	return db
}
