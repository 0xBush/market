package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	HTTPAddr      string
	DatabaseURL   string
	LogLevel      slog.Level
	MigrationsDir string
}

func Load() (Config, error) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	dir := os.Getenv("MIGRATIONS_DIR")
	if dir == "" {
		dir = "migrations"
	}

	level := slog.LevelInfo
	if s := os.Getenv("LOG_LEVEL"); s != "" {
		if err := level.UnmarshalText([]byte(strings.ToLower(s))); err != nil {
			return Config{}, fmt.Errorf("LOG_LEVEL: %w", err)
		}
	}

	return Config{
		HTTPAddr:      addr,
		DatabaseURL:   url,
		LogLevel:      level,
		MigrationsDir: dir,
	}, nil
}
