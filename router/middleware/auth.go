package middleware

import (
	"errors"
	"net/http"
	"strings"

	"uuid"

	"github.com/gin-gonic/gin"
	"github.com/kaeman-dev/kaeman-public-api/jwt"
	"github.com/kaeman-dev/kaeman-public-api/permission"
	"github.com/kaeman-dev/kaeman-public-api/response"
	"github.com/kaeman-dev/kaeman-public-api/server"
	"github.com/kaeman-dev/kaeman-public-api/storage"
	"gorm.io/gorm"
)

func extractToken(value string) string {
	value = strings.TrimSpace(value)
	fields := strings.Fields(value)
	if len(fields) != 2 || !strings.EqualFold(fields[0], "bearer") {
		return value
	}
	return fields[1]
}

func PublicAPIKey(deps *server.Services, optional bool, pms permission.Permission) func(c *gin.Context) error {
	return func(ctx *gin.Context) error {
		value := extractToken(ctx.GetHeader("Authorization"))
		if value == "" {
			if optional {
				return nil
			}
			return response.New("Key required").WithStatus(http.StatusUnauthorized).Write(ctx)
		}
		claims, err := deps.Tokens.ParsePublicAPI(value)
		if err != nil {
			if optional {
				return nil
			}
			return response.New("Invalid public api key").WithStatus(http.StatusUnauthorized).Write(ctx)
		}
		if pms != 0 && *claims.Data.Permissions&pms != pms {
			return response.New("Required permission missing").WithStatus(http.StatusUnauthorized).Write(ctx)
		}
		ctx.Set("publicAPIKey", claims)
		return nil
	}
}

func MinecraftToken(deps *server.Services, optional bool) func(c *gin.Context) error {
	return func(ctx *gin.Context) error {
		value := extractToken(ctx.GetHeader("Authorization"))
		if value == "" {
			if optional {
				return nil
			}
			return response.New("Key required").WithStatus(http.StatusUnauthorized).Write(ctx)
		}
		claims, err := deps.Tokens.ParseMinecraft(value)
		if err != nil {
			if optional {
				return nil
			}
			return response.New("Invalid minecraft key").WithStatus(http.StatusUnauthorized).Write(ctx)
		}

		if publicAPIKey, ok := ctx.Get("publicAPIKey"); ok {
			pub, ok := publicAPIKey.(*jwt.Claims[jwt.PublicAPIClaimsData])
			if !ok {
				return errors.New("unexpected publicAPIKey claims type")
			}
			ownerUID, err := uuid.Parse(pub.Data.UID)
			if err != nil {
				return response.New("Invalid public api key").WithStatus(http.StatusUnauthorized).Write(ctx)
			}
			bound, err := storage.GetIdentity(ctx.Request.Context(), deps.DB, storage.PlatformMinecraft, claims.Data.Minecraft.UUID)
			if errors.Is(err, gorm.ErrRecordNotFound) || (err == nil && bound.UID != ownerUID) {
				return response.New("Minecraft identity not bound to this public api key").WithStatus(http.StatusUnauthorized).Write(ctx)
			}
			if err != nil {
				return err
			}
		}

		ctx.Set("minecraftKey", claims)
		return nil
	}
}
