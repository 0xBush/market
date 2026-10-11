package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v3"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"github.com/0xbush/market/internal/database/sqlc"
	"github.com/0xbush/market/internal/order"
	"github.com/0xbush/market/internal/user"
)

func (s *Server) routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(httplog.RequestLogger(s.log, &httplog.Options{RecoverPanics: true}))

	r.Get("/healthz", s.healthz)
	r.Get("/readyz", s.readyz)
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	q := sqlc.New(s.pool)
	users := user.NewHandler(user.NewService(user.NewRepository(s.pool, q), s.cfg.Auth), s.cfg.Auth, s.log)
	orders := order.NewHandler(order.NewService(order.NewRepository(q)), s.log)

	users.RegisterPublic(r)
	r.Group(func(r chi.Router) {
		r.Use(users.RequireAuth)
		users.RegisterAuthed(r)
		orders.Register(r)
	})

	return r
}
