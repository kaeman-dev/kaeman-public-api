package minecraft

import (
	"github.com/gin-gonic/gin"
	"github.com/kaeman-dev/kaeman-public-api/router/handler/v1/auth/minecraft/api"
	"github.com/kaeman-dev/kaeman-public-api/router/middleware"
	"github.com/kaeman-dev/kaeman-public-api/server"
)

func Register(rg *gin.RouterGroup, deps *server.Services) {
	h := api.Handler{Services: deps}
	rg.POST("/challenge", middleware.RateLimit(deps.Cache), middleware.Wrap(h.Challenge))
	rg.POST("/verify", middleware.RateLimit(deps.Cache), middleware.Wrap(h.Verify))
}
