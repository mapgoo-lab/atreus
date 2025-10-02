package orm

import (
	"testing"
	"time"
)

func TestNewPostgress(t *testing.T) {
	config := &Config{
		DSN:         "host=127.0.0.1 user=root password=123456 dbname=TEST port=5432 sslmode=disable TimeZone=Asia/Shanghai",
		Active:      10,
		Idle:        5,
		IdleTimeout: time.Minute * 5,
	}
	db := NewPostgres(config)
	if db == nil {
		t.Errorf("failed to connect postgres database")
	} else {
		t.Logf("success to connect postgres database")
	}
}
