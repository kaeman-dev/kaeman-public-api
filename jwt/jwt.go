package jwt

import (
	"errors"
	"time"

	standard "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/kaeman-dev/kaeman-public-api/model"
)

type Claims struct {
	Data struct {
		Minecraft   model.MinecraftIdentity `json:"minecraft"`
		Permissions *model.Permission       `json:"permissions,omitempty"`
		Ratelimit   *int                    `json:"ratelimit,omitempty"`
	} `json:"data"`
	standard.RegisteredClaims
}

type Tokens struct{ Secret []byte }

func (t *Tokens) Issue(identity model.MinecraftIdentity, permissions *model.Permission, ratelimit *int, lifetime time.Duration) (string, *Claims, error) {
	now := time.Now().UTC()
	audience := "minecraft-session"
	if permissions != nil {
		audience = "public-api"
	}
	claims := &Claims{RegisteredClaims: standard.RegisteredClaims{ID: uuid.NewString(), Issuer: "kaeman-public-api", Audience: standard.ClaimStrings{audience}, IssuedAt: standard.NewNumericDate(now), ExpiresAt: standard.NewNumericDate(now.Add(lifetime))}}
	claims.Data.Minecraft = identity
	claims.Data.Permissions = permissions
	claims.Data.Ratelimit = ratelimit
	value, err := standard.NewWithClaims(standard.SigningMethodHS256, claims).SignedString(t.Secret)
	return value, claims, err
}

func (t *Tokens) Parse(value, audience string) (*Claims, error) {
	var claims Claims
	token, err := standard.ParseWithClaims(value, &claims, func(*standard.Token) (any, error) { return t.Secret, nil }, standard.WithValidMethods([]string{"HS256"}), standard.WithIssuer("kaeman-public-api"), standard.WithAudience(audience), standard.WithExpirationRequired(), standard.WithIssuedAt())
	if err != nil {
		return nil, err
	}
	validUUID := model.ValidMinecraftUUID(claims.Data.Minecraft.UUID)
	validName := model.ValidMinecraftName(claims.Data.Minecraft.Name)
	if !token.Valid || claims.ID == "" || claims.IssuedAt == nil || claims.ExpiresAt == nil || !claims.ExpiresAt.After(claims.IssuedAt.Time) || !validUUID || !validName {
		return nil, errors.New("invalid JWT identity or lifetime")
	}
	if audience == "public-api" && (claims.Data.Permissions == nil || *claims.Data.Permissions > 2147483647) {
		return nil, errors.New("invalid JWT permissions")
	}
	if audience == "public-api" && claims.Data.Ratelimit != nil && *claims.Data.Ratelimit < 1 {
		return nil, errors.New("invalid JWT ratelimit")
	}
	if audience == "minecraft-session" && claims.Data.Permissions != nil {
		return nil, errors.New("unexpected session permissions")
	}
	return &claims, nil
}
