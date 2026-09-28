package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kaeman-dev/kaeman-public-api/response"
	v1 "github.com/kaeman-dev/kaeman-public-api/router/handler/v1"
	"github.com/kaeman-dev/kaeman-public-api/router/middleware"
	"github.com/kaeman-dev/kaeman-public-api/server"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter(deps *server.Services) *gin.Engine {
	e := gin.New()
	e.Use(middleware.Trace(), middleware.Errors(deps.Log), middleware.Recovery(), middleware.CORS())

	_ = e.SetTrustedProxies(nil)

	e.RedirectTrailingSlash = false
	e.NoRoute(func(c *gin.Context) {
		response.New(http.StatusText(http.StatusNotFound)).WithStatus(http.StatusNotFound).Write(c)
	})

	e.HandleMethodNotAllowed = true
	e.NoMethod(func(c *gin.Context) {
		response.New(http.StatusText(http.StatusMethodNotAllowed)).WithStatus(http.StatusMethodNotAllowed).Write(c)
	})

	if gin.Mode() == gin.DebugMode {
		e.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	v1.Register(e.Group("/v1"), deps)

	return e
}
