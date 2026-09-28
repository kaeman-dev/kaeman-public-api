package storage

import (
	"strconv"
	"strings"
	"uuid"
)

type Platform string

const (
	PlatformKoishi    Platform = "koishi"
	PlatformMinecraft Platform = "minecraft"
)

func NewKaemanUser(nickname string) *KaemanUser {
	return &KaemanUser{Nickname: nickname}
}

type BindFrom string

const (
	BindFromKoishi BindFrom = "koishi"
	BindFromAPI    BindFrom = "api"
)

func NewKoishiIdentity(uid uuid.UUID, koishiID uint64, bindFrom BindFrom) *KaemanIdentity {
	return &KaemanIdentity{
		UID:      uid,
		Platform: PlatformKoishi,
		Identity: strconv.FormatUint(koishiID, 10),
		BindFrom: bindFrom,
	}
}

func NewMinecraftIdentity(uid uuid.UUID, minecraftID uuid.UUID, bindFrom BindFrom) *KaemanIdentity {
	return &KaemanIdentity{
		UID:      uid,
		Platform: PlatformMinecraft,
		Identity: strings.ReplaceAll(minecraftID.String(), "-", ""),
		BindFrom: bindFrom,
	}
}
