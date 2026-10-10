package order

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"market/internal/httputil"
)

type service interface {
	Create(ctx context.Context, in CreateInput) (Order, error)
	Get(ctx context.Context, id int64) (Order, error)
	List(ctx context.Context) ([]Order, error)
}

type Handler struct {
	svc service
	log *slog.Logger
}

func NewHandler(svc service, log *slog.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /orders", h.convert(h.create))
	mux.HandleFunc("GET /orders", h.convert(h.list))
	mux.HandleFunc("GET /orders/{id}", h.convert(h.get))
}

type handlerFunc func(http.ResponseWriter, *http.Request) error

func (h *Handler) convert(fn handlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			var ve *ValidationError
			switch {
			case errors.As(err, &ve):
				httputil.WriteError(w, http.StatusUnprocessableEntity, ve.Error())
			case errors.Is(err, ErrNotFound):
				httputil.WriteError(w, http.StatusNotFound, "order not found")
			default:
				h.log.ErrorContext(r.Context(), "order handler", "err", err)
				httputil.WriteError(w, http.StatusInternalServerError, "internal error")
			}
		}
	}
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) error {
	var req struct {
		Name     string `json:"name"`
		Quantity int32  `json:"quantity"`
	}
	if err := httputil.ReadJSON(w, r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return nil
	}

	o, err := h.svc.Create(r.Context(), CreateInput{Name: req.Name, Quantity: req.Quantity})
	if err != nil {
		return err
	}
	w.Header().Set("Location", fmt.Sprintf("/orders/%d", o.ID))
	httputil.WriteJSON(w, http.StatusCreated, o)
	return nil
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		httputil.WriteError(w, http.StatusBadRequest, "invalid id")
		return nil
	}

	o, err := h.svc.Get(r.Context(), id)
	if err != nil {
		return err
	}
	httputil.WriteJSON(w, http.StatusOK, o)
	return nil
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) error {
	items, err := h.svc.List(r.Context())
	if err != nil {
		return err
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
	return nil
}
