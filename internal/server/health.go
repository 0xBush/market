package server

import (
	"net/http"

	"github.com/0xbush/market/internal/httputil"
)

type HealthBody struct {
	Status string `json:"status"`
}

// healthz reports that the process is up.
//
// @Summary      Liveness
// @Description  Process is up
// @Tags         ops
// @Produce      json
// @Success      200 {object} HealthBody
// @Router       /healthz [get]
func (s *Server) healthz(w http.ResponseWriter, _ *http.Request) {
	httputil.WriteJSON(w, http.StatusOK, HealthBody{Status: "ok"})
}

// readyz reports that Postgres and Redis answer ping.
//
// @Summary      Readiness
// @Description  Postgres and Redis answer ping
// @Tags         ops
// @Produce      json
// @Success      200 {object} HealthBody
// @Failure      503 {object} httputil.ErrorBody
// @Router       /readyz [get]
func (s *Server) readyz(w http.ResponseWriter, r *http.Request) {
	if err := s.pool.Ping(r.Context()); err != nil {
		s.log.ErrorContext(r.Context(), "readyz", "err", err)
		httputil.WriteError(w, http.StatusServiceUnavailable, "not ready")
		return
	}
	if err := s.rdb.Ping(r.Context()).Err(); err != nil {
		s.log.ErrorContext(r.Context(), "readyz", "err", err)
		httputil.WriteError(w, http.StatusServiceUnavailable, "not ready")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, HealthBody{Status: "ok"})
}
