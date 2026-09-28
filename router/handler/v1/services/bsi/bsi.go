package bsi

import (
	"github.com/gin-gonic/gin"
	"github.com/kaeman-dev/kaeman-public-api/permission"
	"github.com/kaeman-dev/kaeman-public-api/router/handler/v1/services/bsi/api"
	"github.com/kaeman-dev/kaeman-public-api/router/middleware"
	"github.com/kaeman-dev/kaeman-public-api/server"
)

func Register(rg *gin.RouterGroup, deps *server.Services) {
	h := api.Handler{Services: deps}
	rg.POST("/queue", middleware.Chain(h.Send,
		middleware.PublicAPIKey(deps, false, permission.SplasherQueue),
		middleware.MinecraftToken(deps, false),
		middleware.RateLimit(deps.RateLimit),
	))

	rg.GET("/gateway", deps.BSI.Listen)
}
