package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Sn0wo2/caelum"
	"github.com/eko/gocache/lib/v4/cache"
	go_cache_store "github.com/eko/gocache/store/go_cache/v4"
	redis_store "github.com/eko/gocache/store/redis/v4"
	"github.com/kaeman-dev/kaeman-public-api/config"
	"github.com/kaeman-dev/kaeman-public-api/gateway"
	"github.com/kaeman-dev/kaeman-public-api/jwt"
	"github.com/kaeman-dev/kaeman-public-api/logging"
	"github.com/kaeman-dev/kaeman-public-api/minecraft"
	"github.com/kaeman-dev/kaeman-public-api/router"
	"github.com/kaeman-dev/kaeman-public-api/server"
	"github.com/kaeman-dev/kaeman-public-api/storage"
	gocache_lib "github.com/patrickmn/go-cache"
	"github.com/redis/go-redis/v9"
	"github.com/ulule/limiter/v3"
	limiter_memstore "github.com/ulule/limiter/v3/drivers/store/memory"
	limiter_redisstore "github.com/ulule/limiter/v3/drivers/store/redis"
)

func main() {
	configPath := flag.String("config", "./data/config.toml", "path to the TOML config file")
	flag.Parse()

	style := caelum.DefaultStyle()
	style.Icons = caelum.IconsNerd

	log := caelum.Init(caelum.Config{
		Targets: []caelum.Target{
			{
				Style: &style,
			},
		},
	})

	logging.RedirectGIN(log.Logger)
	slog.SetDefault(log.Logger)

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

	deps := &server.Services{
		DB:        db,
		Log:       log.Logger,
		Tokens:    &jwt.Token{Secret: []byte(cfg.Auth.JWTSecret)},
		Minecraft: minecraft.New(cfg.Minecraft.MojangSessionBaseURL, http.DefaultClient),
		BSI:       &gateway.Hub{},
	}

	if cfg.Redis.Addr != "" {
		network := "tcp"
		addr := cfg.Redis.Addr
		if path, ok := strings.CutPrefix(addr, "unix://"); ok {
			network, addr = "unix", path
		}
		client := redis.NewClient(&redis.Options{Network: network, Addr: addr, Password: cfg.Redis.Password})
		if err := client.Ping(context.Background()).Err(); err != nil {
			client.Close()
			log.Error("connect to Redis", "error", err)
			os.Exit(1)
		}
		deps.Cache = cache.New[string](redis_store.NewRedis(client))
		limiterStore, err := limiter_redisstore.NewStoreWithOptions(client, limiter.StoreOptions{})
		if err != nil {
			log.Error("initialize ratelimiter", "error", err)
			os.Exit(1)
		}
		deps.RateLimit = limiterStore
	} else {
		deps.Cache = cache.New[string](go_cache_store.NewGoCache(gocache_lib.New(5*time.Minute, 10*time.Minute)))
		deps.RateLimit = limiter_memstore.NewStore()
	}

	httpServer := &http.Server{Addr: cfg.Server.ListenAddr, Handler: router.NewRouter(deps), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16 << 10}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Info("server listening", "address", cfg.Server.ListenAddr)

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- httpServer.ListenAndServe()
	}()

	select {
	case err := <-serveErr:
		log.Error("serve HTTP", "error", err)
		os.Exit(1)
	case <-ctx.Done():
		deps.BSI.Close()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Error("shutdown HTTP", "error", err)
			httpServer.Close()
		}
	}
}
