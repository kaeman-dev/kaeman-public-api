package minecraft

import (
	"github.com/gin-gonic/gin"
	"github.com/kaeman-dev/kaeman-public-api/router/handler/v1/auth/minecraft/api"
	"github.com/kaeman-dev/kaeman-public-api/router/middleware"
	"github.com/kaeman-dev/kaeman-public-api/server"
)

func Register(rg *gin.RouterGroup, deps *server.Services) {
	h := api.Handler{Services: deps}
	rg.GET("/challenge", middleware.Chain(h.Challenge, middleware.RateLimit(deps.RateLimit)))
	rg.POST("/verify", middleware.Chain(h.Verify, middleware.RateLimit(deps.RateLimit)))
}
