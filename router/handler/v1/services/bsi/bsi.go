package bsi

import (
	"github.com/gin-gonic/gin"
	"github.com/kaeman-dev/kaeman-public-api/model"
	"github.com/kaeman-dev/kaeman-public-api/router/handler/v1/services/bsi/api"
	"github.com/kaeman-dev/kaeman-public-api/router/middleware"
	"github.com/kaeman-dev/kaeman-public-api/server"
)

func Register(rg *gin.RouterGroup, deps *server.Services) {
	h := api.Handler{Services: deps}
	rg.POST("/queue", middleware.MinecraftToken(deps.Tokens),
		middleware.OptionalPublicAPIKey(deps.DB, deps.Tokens, middleware.PublicAPIKeyHeader, model.SplashQueue, true),
		middleware.RateLimit(deps.Cache), middleware.Wrap(h.Send))
}
