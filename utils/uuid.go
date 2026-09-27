package utils

import (
	"encoding/hex"
	"strings"

	"github.com/google/uuid"
)

func NewUUID() string {
	id := uuid.New()
	return hex.EncodeToString(id[:])
}

func NormalizeUUID(id string) (string, error) {
	u, err := uuid.Parse(id)
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(u.String(), "-", ""), nil
}
