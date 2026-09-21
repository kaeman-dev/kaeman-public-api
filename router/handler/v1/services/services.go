package services

import (
	"github.com/gin-gonic/gin"
	"github.com/kaeman-dev/kaeman-public-api/router/handler/v1/services/bsi"
	"github.com/kaeman-dev/kaeman-public-api/server"
)

func Register(rg *gin.RouterGroup, deps *server.Services) {
	bsi.Register(rg.Group("/bsi"), deps)
}
