package orm

import (
	"github.com/mapgoo-lab/atreus/pkg/log"
	"gorm.io/gorm/logger"
	"time"
)

type Writer struct{}

func (w *Writer) Printf(format string, args ...interface{}) {
	log.Info(format, args...)
}

func NewLogger() logger.Interface {
	return logger.New(&Writer{}, logger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  logger.Info,
		IgnoreRecordNotFoundError: false,
		Colorful:                  true,
	})
}
