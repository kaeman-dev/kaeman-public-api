package model

import (
	"regexp"
	"strings"

	"github.com/google/uuid"
)

func ValidMinecraftUUID(id string) bool {
	parsed, err := uuid.Parse(id)
	return err == nil && strings.ReplaceAll(parsed.String(), "-", "") == id
}

var minecraftNamePattern = regexp.MustCompile(`^[A-Za-z0-9_]{1,16}$`)

func ValidMinecraftName(name string) bool {
	return minecraftNamePattern.MatchString(name)
}
