//go:build sqlite

package storage

import (
	"strings"

	"github.com/Sn0wo2/agnosco"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func init() {
	RegisterDriver(agnosco.Driver[gorm.Dialector]{
		Name:    "sqlite",
		Schemes: []string{"sqlite", "file"},
		Open: func(dsn string) gorm.Dialector {
			if scheme, rest, ok := strings.Cut(dsn, "://"); ok && strings.EqualFold(scheme, "sqlite") {
				dsn = rest
			}
			return sqlite.Open(dsn)
		},
	})
}
