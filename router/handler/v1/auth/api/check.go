package api

import (
	"github.com/gin-gonic/gin"
	"github.com/kaeman-dev/kaeman-public-api/jwt"
	"github.com/kaeman-dev/kaeman-public-api/response"
)

func Check(c *gin.Context) error {
	claims := c.MustGet("publicJWT").(*jwt.Claims)
	return response.NewResponse("ok", gin.H{"minecraft": claims.Data.Minecraft, "permissions": *claims.Data.Permissions, "expiresAt": claims.ExpiresAt.Time}).Write(c)
}
