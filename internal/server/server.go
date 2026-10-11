package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/0xbush/market/internal/config"
)

type Server struct {
	http *http.Server

	cfg config.Config
	log *slog.Logger

	// Currently, we explicitly declare dependency on these databases for simplicity.
	// Strategy pattern may suit there if the core database might change later.
	pool *pgxpool.Pool

	// Valkey has complete compatibility with Redis lib.
	rdb *redis.Client
}

func New(cfg config.Config, log *slog.Logger, pool *pgxpool.Pool, rdb *redis.Client) *Server {
	s := &Server{cfg: cfg, log: log, pool: pool, rdb: rdb}
	s.http = &http.Server{
		Addr:              cfg.HTTP.Addr,
		Handler:           s.routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	return s
}

func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		s.log.Info("listen", "addr", s.http.Addr)
		errCh <- s.http.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return s.http.Shutdown(shutdownCtx)
	}
}
