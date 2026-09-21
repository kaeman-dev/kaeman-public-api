package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/kaeman-dev/kaeman-public-api/router/handler/v1/auth"
	"github.com/kaeman-dev/kaeman-public-api/router/handler/v1/services"
	"github.com/kaeman-dev/kaeman-public-api/server"
)

func Register(rg *gin.RouterGroup, deps *server.Services) {
	auth.Register(rg.Group("/auth"), deps)
	services.Register(rg.Group("/services"), deps)
}
