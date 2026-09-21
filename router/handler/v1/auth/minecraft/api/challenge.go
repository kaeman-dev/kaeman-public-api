package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kaeman-dev/kaeman-public-api/response"
)

const (
	challengeKeyPrefix = "kaeman:challenge:"
	challengeTTL       = 2 * time.Minute
)

type Challenge struct {
	ServerID  string    `json:"serverID"`
	ExpiresAt time.Time `json:"expiresAt"`
}

func (h Handler) Challenge(c *gin.Context) error {
	serverID := uuid.NewString()

	if err := h.Services.Cache.Set(c.Request.Context(), challengeKeyPrefix+serverID, "", challengeTTL); err != nil {
		return response.NewError(http.StatusServiceUnavailable, "Cannot save challenge")
	}

	c.Header("Cache-Control", "no-store")

	return response.NewData("ok", Challenge{
		ServerID:  serverID,
		ExpiresAt: time.Now().UTC().Add(challengeTTL),
	}).Write(c)
}
