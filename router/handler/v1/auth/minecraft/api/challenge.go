package api

import (
	"time"

	"github.com/eko/gocache/lib/v4/store"
	"github.com/gin-gonic/gin"
	"github.com/kaeman-dev/kaeman-public-api/response"
	"github.com/kaeman-dev/kaeman-public-api/utils"
)

const (
	challengeTTL       = 5 * time.Minute
	challengeKeyPrefix = "kaeman:challenge:"
)

type Challenge struct {
	ServerID  string    `json:"serverID"`
	ExpiresAt time.Time `json:"expiresAt"`
}

func (h Handler) Challenge(c *gin.Context) error {
	serverID := utils.NewUUID()

	if err := h.Services.Cache.Set(c.Request.Context(), challengeKeyPrefix+serverID, "", store.WithExpiration(challengeTTL)); err != nil {
		return err
	}

	c.Header("Cache-Control", "no-store")

	return response.NewData("take your challenge❤", Challenge{
		ServerID:  serverID,
		ExpiresAt: time.Now().UTC().Add(challengeTTL),
	}).Write(c)
}
