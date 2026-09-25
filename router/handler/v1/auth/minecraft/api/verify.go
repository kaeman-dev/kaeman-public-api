package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/eko/gocache/lib/v4/store"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kaeman-dev/kaeman-public-api/model"
	"github.com/kaeman-dev/kaeman-public-api/response"
)

func (h Handler) Verify(c *gin.Context) error {
	var input struct {
		ServerID string `json:"serverID"`
		UUID     string `json:"uuid"`
	}

	if err := response.Decode(c, &input); err != nil {
		return response.NewError(http.StatusBadRequest, err.Error())
	}
	if _, err := uuid.Parse(input.ServerID); err != nil {
		return response.NewError(http.StatusBadRequest, "Invalid serverID")
	}
	if !model.ValidMinecraftUUID(input.UUID) {
		return response.NewError(http.StatusBadRequest, "Invalid UUID")
	}
	_, err := h.Services.Cache.Get(c.Request.Context(), challengeKeyPrefix+input.ServerID)
	if errors.Is(err, store.NotFound{}) {
		return response.NewError(http.StatusUnauthorized, "Invalid or expired challenge")
	}
	if err != nil {
		return response.NewError(http.StatusServiceUnavailable, "Challenge store unavailable")
	}
	identity, err := h.Services.Minecraft.Profile(c.Request.Context(), input.UUID)
	if err != nil {
		return response.NewError(http.StatusServiceUnavailable, "Mojang profile lookup failed")
	}
	verified, err := h.Services.Minecraft.HasJoined(c.Request.Context(), identity.Name, input.ServerID, input.UUID)
	if err != nil {
		return response.NewError(http.StatusServiceUnavailable, "Mojang unavailable")
	}
	if !verified {
		return response.NewError(http.StatusUnauthorized, "Minecraft session verification failed")
	}
	token, claims, err := h.Services.Tokens.Issue(identity, nil, nil, 24*time.Hour)
	if err != nil {
		return response.NewError(http.StatusServiceUnavailable, "Cannot issue token")
	}
	_, err = h.Services.Cache.Get(c.Request.Context(), challengeKeyPrefix+input.ServerID)
	if errors.Is(err, store.NotFound{}) {
		return response.NewError(http.StatusUnauthorized, "Challenge already used or expired")
	}
	if err != nil {
		return response.NewError(http.StatusServiceUnavailable, "Challenge store unavailable")
	}
	if err = h.Services.Cache.Delete(c.Request.Context(), challengeKeyPrefix+input.ServerID); err != nil {
		return response.NewError(http.StatusServiceUnavailable, "Challenge store unavailable")
	}
	c.Header("Cache-Control", "no-store")
	return response.NewData("ok", gin.H{"accessToken": token, "expiresAt": claims.ExpiresAt.Time}).Write(c)
}
