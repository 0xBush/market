package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/0xbush/market/internal/config"
	"github.com/0xbush/market/internal/httputil"
)

const (
	accessCookie  = "access_token"
	refreshCookie = "refresh_token"
)

type ctxKey struct{}

func ContextWithUserID(ctx context.Context, id int64) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(ctxKey{}).(int64)
	return id, ok && id > 0
}

func (h *Handler) RequireAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(accessCookie)
		if err != nil || c.Value == "" {
			httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		id, err := parseAccessToken(c.Value, h.cfg.JWTSecret)
		if err != nil {
			httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		// TODO: denylist access jti until exp

		next.ServeHTTP(w, r.WithContext(ContextWithUserID(r.Context(), id)))
	})
}

func signAccessToken(userID int64, secret string, ttl time.Duration) (string, error) {
	now := time.Now()
	const base2 = 10

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   strconv.FormatInt(userID, base2),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
	})

	return tok.SignedString([]byte(secret))
}

func parseAccessToken(token, secret string) (int64, error) {
	parsed, err := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || !parsed.Valid {
		return 0, ErrUnauthorized
	}

	claims, ok := parsed.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return 0, ErrUnauthorized
	}

	id, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil || id < 1 {
		return 0, ErrUnauthorized
	}
	return id, nil
}

func newRefreshToken() (raw string, hash []byte, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", nil, err
	}
	raw = base64.RawURLEncoding.EncodeToString(b)
	sum := sha256.Sum256([]byte(raw))
	return raw, sum[:], nil
}

func hashRefreshToken(raw string) []byte {
	sum := sha256.Sum256([]byte(raw))
	return sum[:]
}

func setAuthCookies(w http.ResponseWriter, pair TokenPair, cfg config.Auth) {
	setCookie(w, accessCookie, pair.Access, cfg.AccessTTL, cfg.CookieSecure)
	setCookie(w, refreshCookie, pair.Refresh, cfg.RefreshTTL, cfg.CookieSecure)
}

func clearAuthCookies(w http.ResponseWriter, secure bool) {
	clearCookie(w, accessCookie, secure)
	clearCookie(w, refreshCookie, secure)
}

func setCookie(w http.ResponseWriter, name, value string, ttl time.Duration, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   int(ttl.Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearCookie(w http.ResponseWriter, name string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func refreshFromRequest(r *http.Request) (string, error) {
	c, err := r.Cookie(refreshCookie)
	if err != nil || c.Value == "" {
		return "", ErrUnauthorized
	}
	return c.Value, nil
}
