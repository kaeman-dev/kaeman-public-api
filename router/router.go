package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kaeman-dev/kaeman-public-api/response"
	v1 "github.com/kaeman-dev/kaeman-public-api/router/handler/v1"
	"github.com/kaeman-dev/kaeman-public-api/router/middleware"
	"github.com/kaeman-dev/kaeman-public-api/server"
)

func NewRouter(deps *server.Services) *gin.Engine {
	e := gin.New()
	e.Use(middleware.Trace(), middleware.Errors(deps.Log), middleware.Recovery(), middleware.CORS())

	_ = e.SetTrustedProxies(nil)

	e.RedirectTrailingSlash = false
	e.HandleMethodNotAllowed = true
	e.NoRoute(func(c *gin.Context) {
		_ = c.Error(response.NewError(http.StatusNotFound, "not found"))
	})
	e.NoMethod(func(c *gin.Context) {
		_ = c.Error(response.NewError(http.StatusMethodNotAllowed, "method not allowed"))
	})

	v1.Register(e.Group("/v1"), deps)
	e.GET("/ws/bsi", deps.BSI.Listen)
	return e
}
