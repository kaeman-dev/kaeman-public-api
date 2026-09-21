package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kaeman-dev/kaeman-public-api/jwt"
	"github.com/kaeman-dev/kaeman-public-api/model"
	"github.com/kaeman-dev/kaeman-public-api/response"
	"gorm.io/gorm"
)

const PublicAPIKeyHeader = "Public-API-Key"

func extractToken(c *gin.Context, header string) (string, error) {
	value := strings.TrimSpace(c.Request.Header.Get(header))
	if header == "Authorization" {
		fields := strings.Fields(value)
		if len(fields) != 2 || !strings.EqualFold(fields[0], "bearer") {
			return "", response.NewError(http.StatusUnauthorized, "Bearer JWT required")
		}
		value = fields[1]
	}
	if value == "" {
		return "", response.NewError(http.StatusUnauthorized, header+" header required")
	}
	return value, nil
}

func VerifyPublicAPIKey(c *gin.Context, db *gorm.DB, tokens *jwt.Tokens, header string, permission model.Permission, matchUUID string) (*jwt.Claims, error) {
	value, err := extractToken(c, header)
	if err != nil {
		return nil, err
	}
	claims, err := tokens.Parse(value, "public-api")
	if err != nil {
		return nil, response.NewError(http.StatusUnauthorized, "Invalid public JWT")
	}
	var record model.PublicToken
	err = db.WithContext(c.Request.Context()).First(&record, "id = ?", claims.ID).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, response.NewError(http.StatusServiceUnavailable, "Token registry unavailable")
	}
	if err != nil || record.Revoked {
		return nil, response.NewError(http.StatusUnauthorized, "Unknown or revoked token")
	}
	if permission != 0 && *claims.Data.Permissions&permission != permission {
		return nil, response.NewError(http.StatusForbidden, "Required permission missing")
	}
	if matchUUID != "" && claims.Data.Minecraft.UUID != matchUUID {
		return nil, response.NewError(http.StatusForbidden, "Minecraft UUID mismatch")
	}
	return claims, nil
}

func PublicAPIKey(db *gorm.DB, tokens *jwt.Tokens, permission model.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := VerifyPublicAPIKey(c, db, tokens, "Authorization", permission, "")
		if err != nil {
			_ = c.Error(err)
			c.Abort()
			return
		}
		c.Set("publicJWT", claims)
		c.Next()
	}
}

func OptionalPublicAPIKey(db *gorm.DB, tokens *jwt.Tokens, header string, permission model.Permission, matchPlayer bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.TrimSpace(c.GetHeader(header)) == "" {
			c.Next()
			return
		}
		matchUUID := ""
		if matchPlayer {
			matchUUID = c.MustGet("minecraftJWT").(*jwt.Claims).Data.Minecraft.UUID
		}
		claims, err := VerifyPublicAPIKey(c, db, tokens, header, permission, matchUUID)
		if err != nil {
			_ = c.Error(err)
			c.Abort()
			return
		}
		c.Set("publicJWT", claims)
		c.Next()
	}
}

func MinecraftToken(tokens *jwt.Tokens) gin.HandlerFunc {
	return func(c *gin.Context) {
		value, err := extractToken(c, "Authorization")
		if err != nil {
			_ = c.Error(err)
			c.Abort()
			return
		}
		claims, err := tokens.Parse(value, "minecraft-session")
		if err != nil {
			_ = c.Error(response.NewError(http.StatusUnauthorized, "Invalid Minecraft JWT"))
			c.Abort()
			return
		}
		c.Set("minecraftJWT", claims)
		c.Next()
	}
}
