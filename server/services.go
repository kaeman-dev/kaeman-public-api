package server

import (
	"log/slog"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/kaeman-dev/kaeman-public-api/gateway"
	"github.com/kaeman-dev/kaeman-public-api/jwt"
	"github.com/kaeman-dev/kaeman-public-api/minecraft"
	"gorm.io/gorm"
)

type Services struct {
	DB        *gorm.DB
	Log       *slog.Logger
	Tokens    *jwt.Tokens
	Minecraft *minecraft.Client
	Cache     *cache.Cache[string]
	BSI       *gateway.Hub
}
