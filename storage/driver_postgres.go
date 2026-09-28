package storage

import (
	"github.com/Sn0wo2/agnosco"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func init() {
	RegisterDriver(agnosco.Driver[gorm.Dialector]{
		Name:    "postgres",
		Schemes: []string{"postgres", "postgresql"},
		Open:    postgres.Open,
	})
}
