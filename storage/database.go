package storage

import (
	"log/slog"
	"time"

	"github.com/Sn0wo2/agnosco"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var registry = agnosco.NewRegistry[gorm.Dialector]()

func RegisterDriver(driver agnosco.Driver[gorm.Dialector]) {
	registry.Register(driver)
}

func Open(dsn string, log *slog.Logger) (*gorm.DB, error) {
	dialector, err := registry.Open(dsn)
	if err != nil {
		return nil, err
	}
	gormLog := logger.New(slog.NewLogLogger(log.Handler(), slog.LevelError), logger.Config{
		LogLevel:                  logger.Error,
		IgnoreRecordNotFoundError: true,
	})
	db, err := gorm.Open(dialector, &gorm.Config{Logger: gormLog, NowFunc: func() time.Time { return time.Now().UTC() }})
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
