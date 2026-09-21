package main

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sn0wo2/caelum"
	"github.com/kaeman-dev/kaeman-public-api/config"
	"github.com/kaeman-dev/kaeman-public-api/gateway"
	"github.com/kaeman-dev/kaeman-public-api/jwt"
	"github.com/kaeman-dev/kaeman-public-api/minecraft"
	"github.com/kaeman-dev/kaeman-public-api/router"
	"github.com/kaeman-dev/kaeman-public-api/server"
	"github.com/kaeman-dev/kaeman-public-api/storage"
)

func main() {
	configPath := flag.String("config", "./data/config.toml", "path to the TOML config file")
	flag.Parse()

	log := caelum.Init(caelum.Config{})

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Error("load config", "error", err)
		os.Exit(1)
	}
	db, err := storage.Open(cfg.Database.DSN, log.Logger)
	if err != nil {
		log.Error("initialize database", "error", err)
		os.Exit(1)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	deps := &server.Services{DB: db, Log: log.Logger, Tokens: &jwt.Tokens{Secret: []byte(cfg.JWT.Secret)}, Minecraft: minecraft.New(cfg.Minecraft.MojangAPI), BSI: &gateway.Hub{}}

	if cfg.Redis.Addr != "" {
		cache, err := storage.NewRedisKVCache(context.Background(), cfg.Redis.Addr, cfg.Redis.Password)
		if err != nil {
			log.Error("connect to Redis", "error", err)
			os.Exit(1)
		}
		deps.Cache = cache
	} else {
		deps.Cache = storage.NewMemoryKVCache[string]()
	}
	defer deps.Cache.Close()

	httpServer := &http.Server{Addr: cfg.Server.ListenAddr, Handler: router.NewRouter(deps), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16 << 10}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	serverError := make(chan error, 1)
	go func() { serverError <- httpServer.ListenAndServe() }()
	log.Info("server listening", "address", cfg.Server.ListenAddr)
	failed := false
	select {
	case <-ctx.Done():
	case err := <-serverError:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Error("serve HTTP", "error", err)
			failed = true
		}
	}
	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	deps.BSI.Close()
	if err := httpServer.Shutdown(shutdown); err != nil {
		log.Error("shutdown HTTP", "error", err)
		httpServer.Close()
	}
	if failed {
		sqlDB.Close()
		os.Exit(1)
	}
}
