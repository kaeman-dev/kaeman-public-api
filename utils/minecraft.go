package utils

import "regexp"

var minecraftNamePattern = regexp.MustCompile(`^[A-Za-z0-9_]{1,16}$`)

func ValidMinecraftName(name string) bool {
	return minecraftNamePattern.MatchString(name)
}
