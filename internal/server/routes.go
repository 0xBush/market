package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v3"

	"market/internal/database/sqlc"
	"market/internal/order"
)

func (s *Server) routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(httplog.RequestLogger(s.log, &httplog.Options{RecoverPanics: true}))

	r.Get("/healthz", s.healthz)
	r.Get("/readyz", s.readyz)

	orders := order.NewHandler(order.NewService(order.NewRepository(sqlc.New(s.pool))), s.log)
	orders.Register(r)

	return r
}
