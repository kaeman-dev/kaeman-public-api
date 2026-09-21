package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Server    Server    `toml:"server"`
	Database  Database  `toml:"database"`
	Redis     Redis     `toml:"redis"`
	JWT       JWT       `toml:"jwt"`
	Minecraft Minecraft `toml:"minecraft"`
}

type Server struct {
	ListenAddr string `toml:"listen_addr"`
}

type Database struct {
	DSN string `toml:"dsn"`
}

type Redis struct {
	Addr     string `toml:"addr"`
	Password string `toml:"password"`
}

type JWT struct {
	Secret string `toml:"secret"`
}

type Minecraft struct {
	MojangAPI string `toml:"mojang_api"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}
	cfg := &Config{}
	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}
	cfg.ApplyDefaults()
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return cfg, nil
}

func (c *Config) ApplyDefaults() {
	if c.Server.ListenAddr == "" {
		c.Server.ListenAddr = ":38080"
	}
	if c.Database.DSN == "" {
		c.Database.DSN = "./kaeman.db"
	}
	if c.Minecraft.MojangAPI == "" {
		c.Minecraft.MojangAPI = "https://sessionserver.mojang.com"
	}
}

func (c *Config) validate() error {
	var errs error

	if len(c.JWT.Secret) < 32 {
		errs = errors.Join(errors.New("jwt.secret must contain at least 32 bytes"))
	}

	return errs
}
