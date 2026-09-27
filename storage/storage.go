package storage

import (
	"log/slog"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Open(dsn string, log *slog.Logger) (*gorm.DB, error) {
	gormLog := logger.New(slog.NewLogLogger(log.Handler(), slog.LevelError), logger.Config{
		LogLevel:                  logger.Error,
		IgnoreRecordNotFoundError: true,
	})
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: gormLog, NowFunc: func() time.Time { return time.Now().UTC() }})
	if err != nil {
		return nil, err
	}
	connection, err := db.DB()
	if err != nil {
		return nil, err
	}
	connection.SetMaxOpenConns(1)
	return db, nil
}
