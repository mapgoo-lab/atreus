package sql

import (
	"github.com/mapgoo-lab/atreus/pkg/net/netutil/breaker"
	"github.com/mapgoo-lab/atreus/pkg/time"
)

// Config database config.
type Config struct {
	Addr         string          // for trace
	DSN          string          // write data source name.
	ReadDSN      []string        // read data source name.
	Active       int             // pool
	Idle         int             // pool
	IdleTimeout  time.Duration   // connect max life time.
	QueryTimeout time.Duration   // query sql timeout
	ExecTimeout  time.Duration   // execute sql timeout
	TranTimeout  time.Duration   // transaction sql timeout
	Breaker      *breaker.Config // breaker
	DriverName   string			 // Database drivername
}