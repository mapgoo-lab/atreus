package sql

import (
	"github.com/mapgoo-lab/atreus/pkg/log"

	// database driver
	_ "github.com/denisenkom/go-mssqldb"
)

// NewMSSQL new db and retry connection when has error.
func NewMSSQL(c *Config) (db *DB) {
	if c.QueryTimeout == 0 || c.ExecTimeout == 0 || c.TranTimeout == 0 {
		panic("mssql must be set query/execute/transction timeout")
	}

	if len(c.DriverName) == 0 {
		c.DriverName = "sqlserver"
	}

	db, err := Open(c)
	if err != nil {
		log.Error("open mssql error(%v)", err)
		panic(err)
	}
	return
}
