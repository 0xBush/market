package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/0xbush/market/docs/openapi"
	"github.com/0xbush/market/internal/config"
	"github.com/0xbush/market/internal/database"
	"github.com/0xbush/market/internal/server"
)

// @title Market API
// @version 1.0
// @description Marketplace REST API. Access JWT and refresh token are HttpOnly cookies.
// @BasePath /
func main() {
	os.Exit(run())
}

func run() int {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		log.Error("config", "err", err)
		return 1
	}
	
	log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.Log.Level}))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := database.New(ctx, cfg.Database.URL)
	if err != nil {
		log.Error("database", "err", err)
		return 1
	}
	defer pool.Close()

	rdb, err := database.NewRedis(ctx, cfg.Redis.URL)
	if err != nil {
		log.Error("redis", "err", err)
		return 1
	}
	defer rdb.Close()

	if err := server.New(cfg, log, pool, rdb).Run(ctx); err != nil {
		log.Error("server", "err", err)
		return 1
	}
	return 0
}
