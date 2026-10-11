package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"market/internal/httputil"
)

type Server struct {
	log  *slog.Logger
	pool *pgxpool.Pool
	http *http.Server
}

func New(addr string, log *slog.Logger, pool *pgxpool.Pool) *Server {
	s := &Server{log: log, pool: pool}
	s.http = &http.Server{
		Addr:              addr,
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

// healthz reports that the process is up.
//
// @Summary      Liveness
// @Description  Process is up
// @Tags         ops
// @Produce      json
// @Success      200 {object} map[string]string
// @Router       /healthz [get]
func (s *Server) healthz(w http.ResponseWriter, _ *http.Request) {
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// readyz reports that Postgres answers ping.
//
// @Summary      Readiness
// @Description  Postgres answers ping
// @Tags         ops
// @Produce      json
// @Success      200 {object} map[string]string
// @Failure      503 {object} httputil.ErrorBody
// @Router       /readyz [get]
func (s *Server) readyz(w http.ResponseWriter, r *http.Request) {
	if err := s.pool.Ping(r.Context()); err != nil {
		s.log.ErrorContext(r.Context(), "readyz", "err", err)
		httputil.WriteError(w, http.StatusServiceUnavailable, "not ready")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
