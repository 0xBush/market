package order

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

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

func (h *Handler) Register(r chi.Router) {
	r.Post("/orders", h.create)
	r.Get("/orders", h.list)
	r.Get("/orders/{id}", h.get)
}

func (h *Handler) mapErrors(w http.ResponseWriter, r *http.Request, err error) {
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

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Quantity int32  `json:"quantity"`
	}
	if err := httputil.ReadJSON(w, r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	o, err := h.svc.Create(r.Context(), CreateInput{Name: req.Name, Quantity: req.Quantity})
	if err != nil {
		h.mapErrors(w, r, err)
		return
	}
	w.Header().Set("Location", fmt.Sprintf("/orders/%d", o.ID))
	httputil.WriteJSON(w, http.StatusCreated, o)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id < 1 {
		httputil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	o, err := h.svc.Get(r.Context(), id)
	if err != nil {
		h.mapErrors(w, r, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, o)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.List(r.Context())
	if err != nil {
		h.mapErrors(w, r, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}
