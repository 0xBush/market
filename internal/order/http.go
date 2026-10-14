package order

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/0xbush/market/internal/httputil"
)

type service interface {
	Create(ctx context.Context, in CreateInput) (Order, error)
	Get(ctx context.Context, id int64) (Order, error)
	List(ctx context.Context, limit, offset int32) ([]Order, error)
}

type Handler struct {
	svc service
	log *slog.Logger
}

func NewHandler(svc service, log *slog.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

type CreateOrderRequest struct {
	Name     string `json:"name"`
	Quantity int32  `json:"quantity"`
}

type OrderList struct {
	Items []Order `json:"items"`
}

func (h *Handler) Register(r chi.Router) {
	r.Post("/orders", h.create)
	r.Get("/orders", h.list)
	r.Get("/orders/{id}", h.get)
}

func (h *Handler) mapErrors(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrInvalid):
		httputil.WriteError(w, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, ErrNotFound):
		httputil.WriteError(w, http.StatusNotFound, "order not found")
	default:
		h.log.ErrorContext(r.Context(), "order handler", "err", err)
		httputil.WriteError(w, http.StatusInternalServerError, "internal error")
	}
}

// create stores a new order.
//
// @Summary      Create order
// @Description  Create an order
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        request body CreateOrderRequest true "Order to create"
// @Success      201 {object} Order
// @Header       201 {string} Location "/orders/{id}"
// @Failure      400 {object} httputil.ErrorBody
// @Failure      422 {object} httputil.ErrorBody
// @Failure      500 {object} httputil.ErrorBody
// @Router       /orders [post]
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := httputil.ReadJSON(w, r, &req); err != nil {
		httputil.WriteReadError(w, err)
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

// get returns an order by ID.
//
// @Summary      Get order
// @Description  Get an order by ID
// @Tags         orders
// @Produce      json
// @Param        id path int true "Order ID"
// @Success      200 {object} Order
// @Failure      400 {object} httputil.ErrorBody
// @Failure      404 {object} httputil.ErrorBody
// @Failure      500 {object} httputil.ErrorBody
// @Router       /orders/{id} [get]
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
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

func parseID(r *http.Request) (int64, error) {
	const base = 10
	const bitSize = 64
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), base, bitSize)

	return id, err
}

// list returns orders.
//
// @Summary      List orders
// @Description  List orders
// @Tags         orders
// @Produce      json
// @Param        limit  query int false "Page size (1-100, default 50)"
// @Param        offset query int false "Offset (default 0)"
// @Success      200 {object} OrderList
// @Failure      422 {object} httputil.ErrorBody
// @Failure      500 {object} httputil.ErrorBody
// @Router       /orders [get]
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, err := parseQuery(q.Get("limit"))
	if err != nil {
		h.mapErrors(w, r, fmt.Errorf("%w: invalid limit", ErrInvalid))
		return
	}
	offset, err := parseQuery(q.Get("offset"))
	if err != nil {
		h.mapErrors(w, r, fmt.Errorf("%w: invalid offset", ErrInvalid))
		return
	}

	items, err := h.svc.List(r.Context(), limit, offset)
	if err != nil {
		h.mapErrors(w, r, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, OrderList{Items: items})
}

func parseQuery(s string) (int32, error) {
	if s == "" {
		return 0, nil
	}
	v, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(v), nil
}
