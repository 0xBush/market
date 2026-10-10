package server

import (
	"net/http"

	"market/internal/database/sqlc"
	"market/internal/order"
)

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthz)
	mux.HandleFunc("GET /readyz", s.readyz)

	orders := order.NewHandler(order.NewService(order.NewRepository(sqlc.New(s.pool))), s.log)
	orders.Register(mux)

	return recoverer(s.log)(requestLogger(s.log)(requestID(mux)))
}
