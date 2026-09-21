package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/kaeman-dev/kaeman-public-api/router/handler/v1/auth/api"
	"github.com/kaeman-dev/kaeman-public-api/router/handler/v1/auth/minecraft"
	"github.com/kaeman-dev/kaeman-public-api/router/middleware"
	"github.com/kaeman-dev/kaeman-public-api/server"
)

func Register(rg *gin.RouterGroup, deps *server.Services) {
	rg.GET("/check", middleware.PublicAPIKey(deps.DB, deps.Tokens, 0), middleware.RateLimit(deps.Cache), middleware.Wrap(api.Check))
	minecraft.Register(rg.Group("/minecraft"), deps)
}
