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
	Addr string
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

func defaults() Config {
	return Config{
		HTTP:       HTTP{Addr: ":8080"},
		Log:        Log{Level: slog.LevelInfo},
		Migrations: Migrations{Dir: "migrations"},
	}
}

func Load() (Config, error) {
	cfg := defaults()

	envString("HTTP_ADDR", &cfg.HTTP.Addr)
	envString("MIGRATIONS_DIR", &cfg.Migrations.Dir)

	err := errors.Join(
		envRequired("DATABASE_URL", &cfg.Database.URL),
		env("LOG_LEVEL", &cfg.Log.Level, parseLevel),
	)
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func envString(key string, v *string) {
	if s := os.Getenv(key); s != "" {
		*v = s
	}
}

func envRequired(key string, v *string) error {
	s := os.Getenv(key)
	if s == "" {
		return fmt.Errorf("%s is required", key)
	}

	*v = s
	return nil
}

func env[T any](key string, v *T, parse func(string) (T, error)) error {
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

func parseLevel(s string) (slog.Level, error) {
	var level slog.Level
	if err := level.UnmarshalText([]byte(strings.ToLower(s))); err != nil {
		return 0, err
	}
	return level, nil
}
