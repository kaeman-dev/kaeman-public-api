package middleware

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/kaeman-dev/kaeman-public-api/response"
)

func Wrap(h func(c *gin.Context) error) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := h(c); err != nil {
			_ = c.Error(err)
		}
	}
}

func Errors(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		defer func() {
			if len(c.Errors) > 0 {
				causes := make([]error, 0, len(c.Errors))
				for _, entry := range c.Errors {
					causes = append(causes, entry.Err)
				}
				log.ErrorContext(c.Request.Context(), "request failed", "trace_id", requestid.Get(c), "method", c.Request.Method, "path", c.Request.URL.Path, "status", c.Writer.Status(), "error", errors.Join(causes...))
			}
			log.InfoContext(c.Request.Context(), "http request", "trace_id", requestid.Get(c), "method", c.Request.Method, "path", c.Request.URL.Path, "status", c.Writer.Status(), "duration", time.Since(start))
		}()
		c.Next()

		if len(c.Errors) == 0 || c.Writer.Written() {
			return
		}
		response.Handle(c, c.Errors.Last().Err)
	}
}

func Recovery() gin.HandlerFunc {
	return gin.RecoveryWithWriter(io.Discard, func(c *gin.Context, err any) {
		_ = c.Error(fmt.Errorf("panic: %v\n%s", err, debug.Stack()))
	})
}
