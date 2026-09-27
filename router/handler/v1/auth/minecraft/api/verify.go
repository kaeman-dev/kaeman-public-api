package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/eko/gocache/lib/v4/store"
	"github.com/gin-gonic/gin"
	"github.com/kaeman-dev/kaeman-public-api/minecraft"
	"github.com/kaeman-dev/kaeman-public-api/response"
	"github.com/kaeman-dev/kaeman-public-api/utils"
)

type VerifyInput struct {
	ServerID string `json:"serverID"`
	UUID     string `json:"uuid"`
}

const sessionTTL = 24 * time.Hour

func (h Handler) Verify(c *gin.Context) error {
	input := VerifyInput{}

	if err := response.Decode(c, &input); err != nil {
		return response.New("Invalid request body").WithStatus(http.StatusBadRequest).Write(c)
	}
	serverID, err := utils.NormalizeUUID(input.ServerID)
	if err != nil {
		return response.New("Invalid serverID").WithStatus(http.StatusBadRequest).Write(c)
	}
	id, err := utils.NormalizeUUID(input.UUID)
	if err != nil {
		return response.New("Invalid UUID").WithStatus(http.StatusBadRequest).Write(c)
	}

	_, err = h.Services.Cache.Get(c.Request.Context(), challengeKeyPrefix+serverID)
	if errors.Is(err, store.NotFound{}) {
		return response.New("Invalid or expired challenge").WithStatus(http.StatusUnauthorized).Write(c)
	}
	if err != nil {
		return err
	}
	identity, err := h.Services.Minecraft.Profile(c.Request.Context(), id)
	if err != nil {
		return err
	}
	err = h.Services.Minecraft.HasJoined(c.Request.Context(), identity.Name, input.ServerID)
	if errors.Is(err, minecraft.ErrNotJoined) {
		return response.New("Minecraft session verification failed").WithStatus(http.StatusUnauthorized).Write(c)
	}
	if err != nil {
		return err
	}
	token, err := h.Services.Tokens.IssueMinecraft(*identity, sessionTTL)
	if err != nil {
		return err
	}
	_, err = h.Services.Cache.Get(c.Request.Context(), challengeKeyPrefix+serverID)
	if errors.Is(err, store.NotFound{}) {
		return response.New("Challenge already used or expired").WithStatus(http.StatusUnauthorized).Write(c)
	}
	if err != nil {
		return err
	}
	if err = h.Services.Cache.Delete(c.Request.Context(), challengeKeyPrefix+serverID); err != nil {
		return err
	}
	c.Header("Cache-Control", "no-store")
	return response.NewData("ok", gin.H{"accessToken": token, "expiresAt": time.Now().UTC().Add(sessionTTL)}).Write(c)
}
