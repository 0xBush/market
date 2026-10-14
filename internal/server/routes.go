package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog/v3"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"github.com/0xbush/market/internal/database/sqlc"
	"github.com/0xbush/market/internal/order"
)

func (s *Server) routes() http.Handler {
	r := chi.NewRouter()

	if len(s.cfg.HTTP.CORSOrigins) > 0 {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   s.cfg.HTTP.CORSOrigins,
			AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
			AllowCredentials: true,
			MaxAge:           300,
		}))
	}

	r.Use(middleware.RequestID)
	r.Use(httplog.RequestLogger(s.log, &httplog.Options{RecoverPanics: true}))

	r.Get("/healthz", s.healthz)
	r.Get("/readyz", s.readyz)
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	q := sqlc.New(s.pool)
	orders := order.NewHandler(order.NewService(order.NewRepository(q)), s.log)
	orders.Register(r)

	return r
}
