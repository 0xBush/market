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
	"github.com/0xbush/market/internal/user"
)

type service interface {
	Create(ctx context.Context, in CreateInput) (Order, error)
	Get(ctx context.Context, id, userID int64) (Order, error)
	List(ctx context.Context, userID int64) ([]Order, error)
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

func callerID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, ok := user.UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return 0, false
	}
	return id, true
}

// create stores a new order.
//
// @Summary      Create order
// @Description  Create a new order for the current user
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        request body CreateOrderRequest true "Order to create"
// @Success      201 {object} Order
// @Header       201 {string} Location "/orders/{id}"
// @Failure      400 {object} httputil.ErrorBody
// @Failure      401 {object} httputil.ErrorBody
// @Failure      422 {object} httputil.ErrorBody
// @Failure      500 {object} httputil.ErrorBody
// @Router       /orders [post]
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	userID, ok := callerID(w, r)
	if !ok {
		return
	}

	var req CreateOrderRequest
	if err := httputil.ReadJSON(w, r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	o, err := h.svc.Create(r.Context(), CreateInput{UserID: userID, Name: req.Name, Quantity: req.Quantity})
	if err != nil {
		h.mapErrors(w, r, err)
		return
	}
	w.Header().Set("Location", fmt.Sprintf("/orders/%d", o.ID))
	httputil.WriteJSON(w, http.StatusCreated, o)
}

// get returns one of the current user's orders.
//
// @Summary      Get order
// @Description  Get an order by ID belonging to the current user
// @Tags         orders
// @Produce      json
// @Param        id path int true "Order ID"
// @Success      200 {object} Order
// @Failure      400 {object} httputil.ErrorBody
// @Failure      401 {object} httputil.ErrorBody
// @Failure      404 {object} httputil.ErrorBody
// @Failure      500 {object} httputil.ErrorBody
// @Router       /orders/{id} [get]
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	userID, ok := callerID(w, r)
	if !ok {
		return
	}

	id, err := parseID(r)
	if err != nil || id < 1 {
		httputil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	o, err := h.svc.Get(r.Context(), id, userID)
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

// list returns the current user's orders.
//
// @Summary      List orders
// @Description  List orders belonging to the current user
// @Tags         orders
// @Produce      json
// @Success      200 {object} OrderList
// @Failure      401 {object} httputil.ErrorBody
// @Failure      500 {object} httputil.ErrorBody
// @Router       /orders [get]
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	userID, ok := callerID(w, r)
	if !ok {
		return
	}

	items, err := h.svc.List(r.Context(), userID)
	if err != nil {
		h.mapErrors(w, r, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, OrderList{Items: items})
}
