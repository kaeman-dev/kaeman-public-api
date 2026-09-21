package server

import (
	"log/slog"

	"github.com/kaeman-dev/kaeman-public-api/gateway"
	"github.com/kaeman-dev/kaeman-public-api/jwt"
	"github.com/kaeman-dev/kaeman-public-api/minecraft"
	"github.com/kaeman-dev/kaeman-public-api/storage"
	"gorm.io/gorm"
)

type Services struct {
	DB        *gorm.DB
	Log       *slog.Logger
	Tokens    *jwt.Tokens
	Minecraft *minecraft.Client
	Cache     storage.KVCache[string]
	BSI       *gateway.Hub
}
