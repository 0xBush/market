package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/0xbush/market/internal/config"
	"github.com/0xbush/market/internal/httputil"
	"github.com/0xbush/market/internal/user"
)

type service interface {
	Register(ctx context.Context, in user.Credentials) (user.User, TokenPair, error)
	Login(ctx context.Context, in user.Credentials) (user.User, TokenPair, error)
	Refresh(ctx context.Context, rawRefresh string) (TokenPair, error)
	Logout(ctx context.Context, rawRefresh string) error
	Me(ctx context.Context, userID int64) (user.User, error)
}

type Handler struct {
	svc service
	cfg config.Auth
	log *slog.Logger
}

func NewHandler(svc service, cfg config.Auth, log *slog.Logger) *Handler {
	return &Handler{svc: svc, cfg: cfg, log: log}
}

type credentialsRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) RegisterPublic(r chi.Router) {
	r.Post("/register", h.register)
	r.Post("/login", h.login)
	r.Post("/refresh", h.refresh)
	r.Post("/logout", h.logout)
}

func (h *Handler) RegisterAuthed(r chi.Router) {
	r.Get("/me", h.me)
}

func (h *Handler) mapErrors(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, user.ErrInvalid):
		httputil.WriteError(w, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, user.ErrConflict):
		httputil.WriteError(w, http.StatusConflict, "email already taken")
	case errors.Is(err, ErrUnauthorized):
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
	default:
		h.log.ErrorContext(r.Context(), "auth handler", "err", err)
		httputil.WriteError(w, http.StatusInternalServerError, "internal error")
	}
}

func (h *Handler) writeAuth(w http.ResponseWriter, status int, u user.User, pair TokenPair) {
	setAuthCookies(w, pair, h.cfg)
	httputil.WriteJSON(w, status, u)
}

// register creates an account and signs the caller in.
//
// @Summary      Register
// @Description  Create an account and set access/refresh cookies
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body credentialsRequest true "Credentials"
// @Success      201 {object} user.User
// @Failure      400 {object} httputil.ErrorBody
// @Failure      409 {object} httputil.ErrorBody
// @Failure      422 {object} httputil.ErrorBody
// @Failure      500 {object} httputil.ErrorBody
// @Router       /register [post]
func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req credentialsRequest
	if err := httputil.ReadJSON(w, r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	u, pair, err := h.svc.Register(r.Context(), user.Credentials{Email: req.Email, Password: req.Password})
	if err != nil {
		h.mapErrors(w, r, err)
		return
	}
	h.writeAuth(w, http.StatusCreated, u, pair)
}

// login signs the caller in.
//
// @Summary      Login
// @Description  Set access/refresh cookies
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body credentialsRequest true "Credentials"
// @Success      200 {object} user.User
// @Failure      400 {object} httputil.ErrorBody
// @Failure      401 {object} httputil.ErrorBody
// @Failure      422 {object} httputil.ErrorBody
// @Failure      500 {object} httputil.ErrorBody
// @Router       /login [post]
func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req credentialsRequest
	if err := httputil.ReadJSON(w, r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	u, pair, err := h.svc.Login(r.Context(), user.Credentials{Email: req.Email, Password: req.Password})
	if err != nil {
		if errors.Is(err, ErrUnauthorized) {
			httputil.WriteError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		h.mapErrors(w, r, err)
		return
	}
	h.writeAuth(w, http.StatusOK, u, pair)
}

// refresh rotates the refresh token and issues a new access JWT.
//
// @Summary      Refresh
// @Description  Rotate refresh cookie and set a new access cookie
// @Tags         auth
// @Success      204
// @Failure      401 {object} httputil.ErrorBody
// @Failure      500 {object} httputil.ErrorBody
// @Router       /refresh [post]
func (h *Handler) refresh(w http.ResponseWriter, r *http.Request) {
	raw, err := refreshFromRequest(r)
	if err != nil {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	pair, err := h.svc.Refresh(r.Context(), raw)
	if err != nil {
		h.mapErrors(w, r, err)
		return
	}
	setAuthCookies(w, pair, h.cfg)
	w.WriteHeader(http.StatusNoContent)
}

// logout revokes the refresh token and clears cookies.
//
// @Summary      Logout
// @Description  Revoke refresh token and clear cookies
// @Tags         auth
// @Success      204
// @Failure      500 {object} httputil.ErrorBody
// @Router       /logout [post]
func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	raw, _ := refreshFromRequest(r)
	if err := h.svc.Logout(r.Context(), raw); err != nil {
		h.mapErrors(w, r, err)
		return
	}
	clearAuthCookies(w, h.cfg.CookieSecure)
	w.WriteHeader(http.StatusNoContent)
}

// me returns the current user.
//
// @Summary      Current user
// @Description  Return the account for the access cookie
// @Tags         auth
// @Produce      json
// @Success      200 {object} user.User
// @Failure      401 {object} httputil.ErrorBody
// @Failure      500 {object} httputil.ErrorBody
// @Router       /me [get]
func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	id, ok := UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	u, err := h.svc.Me(r.Context(), id)
	if err != nil {
		h.mapErrors(w, r, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, u)
}
