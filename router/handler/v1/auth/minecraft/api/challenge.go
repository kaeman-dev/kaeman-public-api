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
	// ServerID 一次性挑战 ID，小写无横杠 UUID，签发后 5 分钟内有效且只能消费一次
	ServerID string `json:"serverID" example:"0f8f8c8e5d3c4b2a9e7d6c5b4a3f2e1d"`
	// ExpiresAt 挑战过期时间
	ExpiresAt time.Time `json:"expiresAt" example:"2026-09-29T12:05:00Z"`
}

// Challenge godoc
// @Summary 获取 Minecraft 会话验证挑战
// @Description 生成一个 5 分钟内有效的 serverID 挑战，客户端需让玩家在 Mojang 会话服务器上把该 serverID join 到指定服务器，再调用 /verify 完成验证。挑战为一次性，verify 成功后立即失效。
// @Tags auth
// @Success 200 {object} response.Response[api.Challenge]
// @Header 200 {string} Cache-Control "no-store"
// @Failure 429 {object} response.Response[any] "Too Many Requests"
// @Failure 500 {object} response.Response[any]
// @Router /v1/auth/minecraft/challenge [get]
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
