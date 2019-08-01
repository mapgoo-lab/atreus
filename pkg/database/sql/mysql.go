package sql

import (
	"github.com/mapgoo-lab/atreus/pkg/log"

	// database driver
	_ "github.com/go-sql-driver/mysql"
)

// NewMySQL new db and retry connection when has error.
func NewMySQL(c *Config) (db *DB) {
	if c.QueryTimeout == 0 || c.ExecTimeout == 0 || c.TranTimeout == 0 {
		panic("mysql must be set query/execute/transction timeout")
	}

	if len(c.DriverName) == 0 {
		c.DriverName = "mysql"
	}

	db, err := Open(c)
	if err != nil {
		log.Error("open mysql error(%v)", err)
		panic(err)
	}
	return
}
