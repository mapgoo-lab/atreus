package orm

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func NewPostgres(config *Config) *gorm.DB {
	if config.GormConfig == nil {
		config.GormConfig = &gorm.Config{
			NamingStrategy: schema.NamingStrategy{
				TablePrefix:   "",
				SingularTable: false,
				NoLowerCase:   true, // 不自动转换为小写
			},
		}
	}
  
	if config.GormConfig.Logger == nil {
		config.GormConfig.Logger = NewLogger()
	}

	db, err := gorm.Open(postgres.Open(config.DSN), config.GormConfig)
	if err != nil {
		return nil
	}

	//连接池
	sqlDb, err := db.DB()
	if err != nil {
		return nil
	}
	sqlDb.SetMaxOpenConns(config.Active)
	sqlDb.SetMaxIdleConns(config.Idle)
	sqlDb.SetConnMaxLifetime(config.IdleTimeout)

	err = db.Use(&ObsPlugin{})
	if err != nil {
		return nil
	}

	return db
}
