//go:build clickhouse

package storage

import (
	"github.com/Sn0wo2/agnosco"
	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
)

func init() {
	RegisterDriver(agnosco.Driver[gorm.Dialector]{
		Name:    "clickhouse",
		Schemes: []string{"clickhouse"},
		Open:    clickhouse.Open,
	})
}
