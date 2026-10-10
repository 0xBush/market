package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"market/internal/config"
	"market/internal/database"
	"market/internal/server"
)

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
	log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := database.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("database", "err", err)
		return 1
	}
	defer pool.Close()

	if err := server.New(cfg.HTTPAddr, log, pool).Run(ctx); err != nil {
		log.Error("server", "err", err)
		return 1
	}
	return 0
}
