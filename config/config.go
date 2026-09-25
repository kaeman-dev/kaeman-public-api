package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"reflect"
	"strings"

	"github.com/Sn0wo2/ordo"
	"github.com/Sn0wo2/ordo/format"
	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Server    Server    `toml:"server"`
	Database  Database  `toml:"database"`
	Redis     Redis     `toml:"redis,omitempty"`
	Auth      Auth      `toml:"auth"`
	Minecraft Minecraft `toml:"minecraft"`
}

type Server struct {
	ListenAddr   string   `toml:"listen_addr"`
	TrustProxies []string `toml:"trust_proxies,omitempty"`
}

type Database struct {
	DSN string `toml:"dsn"`
}

type Redis struct {
	Addr     string `toml:"addr,omitempty"`
	Password string `toml:"password,omitempty"`
}

type Auth struct {
	JWTSecret string `toml:"secret"`
}

type Minecraft struct {
	MojangAPI string `toml:"mojang_api"`
}

func Load(path string) (*Config, error) {
	loader := &ordo.Loader[Config]{
		Formats: []format.Format[Config]{format.NewFormatter[Config](
			"toml", []string{".toml"}, 10,
			toml.Unmarshal,
			toml.Marshal,
		)},
		Default: &Config{
			Server: Server{
				ListenAddr: ":38080",
			},
			Database: Database{
				DSN: "./data/kaeman.db",
			},
			Auth: Auth{
				JWTSecret: "kaeman",
			},
			Minecraft: Minecraft{
				MojangAPI: "https://api.mojang.com",
			},
		},
		Validate: func(cfg *Config) error {
			var errs error

			errs = errors.Join(errs, cfg.Validate())

			for _, proxy := range cfg.Server.TrustProxies {
				if net.ParseIP(proxy) == nil {
					if _, _, err := net.ParseCIDR(proxy); err != nil {
						errs = errors.Join(errs, fmt.Errorf("server.trust_proxies contains invalid IP or CIDR: %q", proxy))
					}
				}
			}
			if len(cfg.Auth.JWTSecret) > 0 && len(cfg.Auth.JWTSecret) < 32 {
				errs = errors.Join(errs, errors.New("auth.secret must contain at least 32 bytes"))
			}
			if cfg.Minecraft.MojangAPI != "" {
				if u, err := url.Parse(cfg.Minecraft.MojangAPI); err != nil || u.Scheme == "" || u.Host == "" {
					errs = errors.Join(errs, fmt.Errorf("minecraft.mojang_api must be a valid URL: %q", cfg.Minecraft.MojangAPI))
				}
			}

			return errs
		},
	}

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		if err := loader.Save(loader.Default, path); err != nil {
			return nil, fmt.Errorf("failed to create default config file %s: %w", path, err)
		}
		return nil, fmt.Errorf("default config file created at %s, please edit it and restart", path)
	}

	cfg, _, err := loader.Load(path)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (cfg *Config) Validate() error {
	if cfg == nil {
		return errors.New("config must not be nil")
	}

	var errs error

	var walk func(value reflect.Value, path string)
	walk = func(value reflect.Value, path string) {
		t := value.Type()
		for i := range t.NumField() {
			field := t.Field(i)
			if !field.IsExported() {
				continue
			}

			tag := field.Tag.Get("toml")
			if tag == "-" {
				continue
			}

			name, opts, _ := strings.Cut(tag, ",")
			if strings.Contains(opts, "omitempty") || strings.Contains(opts, "omitzero") {
				continue
			}
			if name == "" {
				name = field.Name
			}

			child := value.Field(i)
			childPath := name
			if path != "" {
				childPath = path + "." + name
			}

			if child.Kind() == reflect.Struct {
				walk(child, childPath)
				continue
			}
			if child.IsZero() {
				errs = errors.Join(errs, fmt.Errorf("%s is required", childPath))
			}
		}
	}

	walk(reflect.ValueOf(cfg).Elem(), "")

	return errs
}
