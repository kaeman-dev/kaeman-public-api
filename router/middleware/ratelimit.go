package middleware

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
	"github.com/gin-gonic/gin"
	"github.com/kaeman-dev/kaeman-public-api/jwt"
	"github.com/kaeman-dev/kaeman-public-api/response"
)

const (
	DefaultRatelimit   = 60
	RatelimitIdleTTL   = 5 * time.Minute
	ratelimitKeyPrefix = "kaeman:ratelimit:"
)

type ratelimitState struct {
	Tokens float64 `json:"tokens"`
	Last   int64   `json:"last"`
}

var ratelimitMu sync.Mutex

func RateLimit(cache *cache.Cache[string]) func(c *gin.Context) error {
	return func(c *gin.Context) error {
		count := DefaultRatelimit
		key := ratelimitKeyPrefix + c.FullPath() + ":ip:" + c.ClientIP()
		if value, ok := c.Get("publicAPIKey"); ok {
			if claims, ok := value.(*jwt.Claims[jwt.PublicAPIClaimsData]); ok {
				key = ratelimitKeyPrefix + c.FullPath() + ":key:" + claims.ID + ":" + c.ClientIP()
				if claims.Data.Ratelimit != nil && *claims.Data.Ratelimit > 0 {
					count = *claims.Data.Ratelimit
				}
			}
		} else if value, ok := c.Get("minecraftKey"); ok {
			if claims, ok := value.(*jwt.Claims[jwt.MinecraftClaimsData]); ok {
				key = ratelimitKeyPrefix + c.FullPath() + ":key:" + claims.ID + ":" + c.ClientIP()
			}
		}
		burst := float64(count)
		refill := burst / 60
		now := time.Now()
		ctx := c.Request.Context()
		ratelimitMu.Lock()
		state := ratelimitState{Tokens: burst, Last: now.UnixNano()}
		if raw, err := cache.Get(ctx, key); err == nil {
			if err := json.Unmarshal([]byte(raw), &state); err != nil || state.Tokens < 0 || state.Tokens > burst {
				state = ratelimitState{Tokens: burst, Last: now.UnixNano()}
			} else {
				elapsed := now.Sub(time.Unix(0, state.Last)).Seconds()
				if elapsed < 0 {
					elapsed = 0
				}
				state.Tokens = math.Min(burst, state.Tokens+elapsed*refill)
				state.Last = now.UnixNano()
			}
		}
		retry := 0
		if state.Tokens >= 1 {
			state.Tokens--
		} else {
			retry = int(math.Ceil((1 - state.Tokens) / refill))
			if retry < 1 {
				retry = 1
			}
		}
		if encoded, err := json.Marshal(state); err == nil {
			_ = cache.Set(ctx, key, string(encoded), store.WithExpiration(RatelimitIdleTTL))
		}
		ratelimitMu.Unlock()
		if retry > 0 {
			c.Header("Retry-After", strconv.Itoa(retry))
			return response.Status(http.StatusTooManyRequests).Write(c)
		}
		return nil
	}
}
