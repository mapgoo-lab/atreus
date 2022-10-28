package orm

import (
	"github.com/mapgoo-lab/atreus/pkg/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewMysql(dsn string, config *gorm.Config) *gorm.DB {
	if config.Logger == nil {
		config.Logger = NewLogger()
	}

	db, err := gorm.Open(mysql.Open(dsn), config)
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
