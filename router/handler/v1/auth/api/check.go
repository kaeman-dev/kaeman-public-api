package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kaeman-dev/kaeman-public-api/jwt"
	"github.com/kaeman-dev/kaeman-public-api/response"
	"github.com/kaeman-dev/kaeman-public-api/server"
	"github.com/kaeman-dev/kaeman-public-api/storage"
	"uuid"
)

type Handler struct{ Services *server.Services }

func (h Handler) Check(c *gin.Context) error {
	claims := c.MustGet("publicAPIKey").(*jwt.Claims[jwt.PublicAPIClaimsData])
	uid, err := uuid.Parse(claims.Data.UID)
	if err != nil {
		return response.New("Invalid public api key").WithStatus(http.StatusUnauthorized).Write(c)
	}
	identities, err := storage.ListIdentitiesByUID(c.Request.Context(), h.Services.DB, uid)
	if err != nil {
		return err
	}
	minecrafts := make([]string, 0, len(identities))
	for _, identity := range identities {
		if identity.Platform == storage.PlatformMinecraft {
			minecrafts = append(minecrafts, identity.Identity)
		}
	}
	return response.NewData("ok",
		gin.H{"minecrafts": minecrafts, "permissions": *claims.Data.Permissions, "expiresAt": claims.ExpiresAt.Time}).Write(c)
}
