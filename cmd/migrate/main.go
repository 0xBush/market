package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"market/internal/config"
	"market/internal/database"
)

func main() {
	os.Exit(run())
}

func run() int {
	down := flag.Bool("down", false, "roll back one migration")
	flag.Parse()

	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		log.Error("config", "err", err)
		return 1
	}

	ctx := context.Background()
	pool, err := database.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("database", "err", err)
		return 1
	}
	defer pool.Close()

	if *down {
		if err := database.MigrateDown(ctx, pool, cfg.MigrationsDir); err != nil {
			log.Error("migrate down", "err", err)
			return 1
		}
		log.Info("migrated down one version")
		return 0
	}

	if err := database.MigrateUp(ctx, pool, cfg.MigrationsDir); err != nil {
		log.Error("migrate up", "err", err)
		return 1
	}
	log.Info("migrated up")
	return 0
}
