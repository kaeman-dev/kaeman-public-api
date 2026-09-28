//go:build sqlserver

package storage

import (
	"github.com/Sn0wo2/agnosco"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

func init() {
	RegisterDriver(agnosco.Driver[gorm.Dialector]{
		Name:    "sqlserver",
		Schemes: []string{"sqlserver"},
		Open:    sqlserver.Open,
	})
}
