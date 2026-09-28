//go:build mysql

package storage

import (
	"strings"

	"github.com/Sn0wo2/agnosco"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func init() {
	RegisterDriver(agnosco.Driver[gorm.Dialector]{
		Name:    "mysql",
		Schemes: []string{"mysql"},
		Open: func(dsn string) gorm.Dialector {
			_, rest, _ := strings.Cut(dsn, "://")
			return mysql.Open(rest)
		},
	})
}
