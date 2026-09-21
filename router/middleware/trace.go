package middleware

import (
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

func Trace() gin.HandlerFunc {
	return requestid.New(
		requestid.WithCustomHeaderStrKey("Trace-ID"),
	)
}
