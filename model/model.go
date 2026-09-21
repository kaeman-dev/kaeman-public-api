package model

import "time"

type Permission uint64

const SplashQueue Permission = 1 << iota

type MinecraftIdentity struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}
type PublicToken struct {
	ID      string `gorm:"primaryKey"`
	Revoked bool
}
type Event struct {
	ID        string `gorm:"primaryKey"`
	Payload   []byte
	CreatedAt time.Time
}
