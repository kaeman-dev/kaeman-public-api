package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kaeman-dev/kaeman-public-api/jwt"
	"github.com/kaeman-dev/kaeman-public-api/response"
	"github.com/ulule/limiter/v3"
)

const ratelimitKeyPrefix = "kaeman:ratelimit:"

func RateLimit(store limiter.Store) func(c *gin.Context) error {
	return func(c *gin.Context) error {
		count := int64(60)
		key := ratelimitKeyPrefix + c.FullPath() + ":ip:" + c.ClientIP()
		if value, ok := c.Get("publicAPIKey"); ok {
			if claims, ok := value.(*jwt.Claims[jwt.PublicAPIClaimsData]); ok {
				key = ratelimitKeyPrefix + c.FullPath() + ":key:" + claims.ID + ":" + c.ClientIP()
				if claims.Data.Ratelimit != nil && *claims.Data.Ratelimit > 0 {
					count = int64(*claims.Data.Ratelimit)
				}
			}
		} else if value, ok := c.Get("minecraftKey"); ok {
			if claims, ok := value.(*jwt.Claims[jwt.MinecraftClaimsData]); ok {
				key = ratelimitKeyPrefix + c.FullPath() + ":key:" + claims.ID + ":" + c.ClientIP()
			}
		}
		state, err := store.Get(c.Request.Context(), key, limiter.Rate{Period: time.Minute, Limit: count})
		if err != nil {
			return err
		}
		resetIn := state.Reset - time.Now().Unix()
		if resetIn < 0 {
			resetIn = 0
		}
		limit := strconv.FormatInt(state.Limit, 10)
		remaining := strconv.FormatInt(state.Remaining, 10)
		c.Header("RateLimit-Limit", limit)
		c.Header("RateLimit-Remaining", remaining)
		c.Header("RateLimit-Reset", strconv.FormatInt(resetIn, 10))
		c.Header("X-RateLimit-Limit", limit)
		c.Header("X-RateLimit-Remaining", remaining)
		c.Header("X-RateLimit-Reset", strconv.FormatInt(state.Reset, 10))
		if state.Reached {
			c.Header("Retry-After", strconv.FormatInt(resetIn, 10))
			return response.Status(http.StatusTooManyRequests).Write(c)
		}
		return nil
	}
}
