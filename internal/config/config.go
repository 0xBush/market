package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	HTTP       HTTP
	Database   Database
	Log        Log
	Migrations Migrations
}

type HTTP struct {
	Addr        string
	CORSOrigins []string
}

type Database struct {
	URL string
}

type Log struct {
	Level slog.Level
}

type Migrations struct {
	Dir string
}

func Load() (Config, error) {
	var cfg Config

	cfg.HTTP.Addr = env("HTTP_ADDR", ":8080")
	cfg.HTTP.CORSOrigins = []string{"http://localhost:5173", "http://localhost:3000"}
	cfg.Migrations.Dir = env("MIGRATIONS_DIR", "migrations")
	cfg.Log.Level = slog.LevelInfo

	err := errors.Join(
		envRequired("DATABASE_URL", &cfg.Database.URL),
		envParse("CORS_ORIGINS", &cfg.HTTP.CORSOrigins, parseComma),
		envParse("LOG_LEVEL", &cfg.Log.Level, parseLogging),
	)
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func env(key, defaults string) string {
	if s := os.Getenv(key); s != "" {
		return s
	}

	return defaults
}

func envRequired(key string, v *string) error {
	s := os.Getenv(key)
	if s == "" {
		return fmt.Errorf("%s is required", key)
	}

	*v = s
	return nil
}

func envParse[T any](key string, v *T, parse func(string) (T, error)) error {
	s := os.Getenv(key)
	if s == "" {
		return nil
	}

	x, err := parse(s)
	if err != nil {
		return fmt.Errorf("%s: %w", key, err)
	}

	*v = x
	return nil
}

func parseComma(s string) ([]string, error) {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out, nil
}

func parseLogging(s string) (slog.Level, error) {
	var level slog.Level
	if err := level.UnmarshalText([]byte(strings.ToLower(s))); err != nil {
		return 0, err
	}
	return level, nil
}
