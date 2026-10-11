package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"

	"github.com/0xbush/market/internal/config"
)

const (
	bcryptCost    = 12
	minPassword   = 8
	maxPassword   = 72
	maxEmailRunes = 254
)

type repo interface {
	Create(ctx context.Context, email, passwordHash string) (User, error)
	GetByEmail(ctx context.Context, email string) (User, string, error)
	GetByID(ctx context.Context, id int64) (User, error)
	CreateRefresh(ctx context.Context, userID int64, tokenHash []byte, expiresAt time.Time) error
	DeleteRefresh(ctx context.Context, tokenHash []byte) error
	RotateRefresh(ctx context.Context, oldHash, newHash []byte, expiresAt time.Time) (int64, error)
}

type Service struct {
	repo repo
	auth config.Auth
}

func NewService(r repo, auth config.Auth) *Service {
	return &Service{repo: r, auth: auth}
}

func (s *Service) Register(ctx context.Context, in Credentials) (User, TokenPair, error) {
	email, password, err := validateCredentials(in)
	if err != nil {
		return User{}, TokenPair{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return User{}, TokenPair{}, err
	}

	u, err := s.repo.Create(ctx, email, string(hash))
	if err != nil {
		return User{}, TokenPair{}, err
	}

	pair, err := s.issueTokens(ctx, u.ID)
	if err != nil {
		return User{}, TokenPair{}, err
	}
	return u, pair, nil
}

func (s *Service) Login(ctx context.Context, in Credentials) (User, TokenPair, error) {
	email, password, err := validateCredentials(in)
	if err != nil {
		return User{}, TokenPair{}, err
	}

	u, passwordHash, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
			return User{}, TokenPair{}, ErrUnauthorized
		}
		return User{}, TokenPair{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return User{}, TokenPair{}, ErrUnauthorized
	}

	pair, err := s.issueTokens(ctx, u.ID)
	if err != nil {
		return User{}, TokenPair{}, err
	}
	return u, pair, nil
}

func (s *Service) Refresh(ctx context.Context, rawRefresh string) (TokenPair, error) {
	if rawRefresh == "" {
		return TokenPair{}, ErrUnauthorized
	}

	raw, hash, err := newRefreshToken()
	if err != nil {
		return TokenPair{}, err
	}

	userID, err := s.repo.RotateRefresh(ctx, hashRefreshToken(rawRefresh), hash, time.Now().Add(s.auth.RefreshTTL))
	if err != nil {
		return TokenPair{}, err
	}

	access, err := signAccessToken(userID, s.auth.JWTSecret, s.auth.AccessTTL)
	if err != nil {
		return TokenPair{}, err
	}
	return TokenPair{Access: access, Refresh: raw}, nil
}

func (s *Service) Logout(ctx context.Context, rawRefresh string) error {
	if rawRefresh == "" {
		return nil
	}
	return s.repo.DeleteRefresh(ctx, hashRefreshToken(rawRefresh))
}

func (s *Service) Me(ctx context.Context, userID int64) (User, error) {
	u, err := s.repo.GetByID(ctx, userID)
	if errors.Is(err, ErrNotFound) {
		return User{}, ErrUnauthorized
	}
	return u, err
}

func (s *Service) issueTokens(ctx context.Context, userID int64) (TokenPair, error) {
	access, err := signAccessToken(userID, s.auth.JWTSecret, s.auth.AccessTTL)
	if err != nil {
		return TokenPair{}, err
	}
	raw, hash, err := newRefreshToken()
	if err != nil {
		return TokenPair{}, err
	}
	if err := s.repo.CreateRefresh(ctx, userID, hash, time.Now().Add(s.auth.RefreshTTL)); err != nil {
		return TokenPair{}, err
	}
	return TokenPair{Access: access, Refresh: raw}, nil
}

func validateCredentials(in Credentials) (string, string, error) {
	email, err := normalizeEmail(in.Email)
	if err != nil {
		return "", "", err
	}
	if n := utf8.RuneCountInString(in.Password); n < minPassword || n > maxPassword {
		return "", "", fmt.Errorf("%w: password must be %d-%d characters", ErrInvalid, minPassword, maxPassword)
	}
	return email, in.Password, nil
}

func normalizeEmail(s string) (string, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	n := utf8.RuneCountInString(s)
	if n < 3 || n > maxEmailRunes {
		return "", fmt.Errorf("%w: invalid email", ErrInvalid)
	}
	at := strings.IndexByte(s, '@')
	if at < 1 || at != strings.LastIndexByte(s, '@') || at == len(s)-1 {
		return "", fmt.Errorf("%w: invalid email", ErrInvalid)
	}
	return s, nil
}

// dummyHash is a valid bcrypt hash used when the email is unknown so login
// timing is closer to the password-mismatch path.
var dummyHash = []byte("$2a$12$R9h/cIPz0gi.URNNX3kh2OPST9/PgBkqquzi.Ss7KIUgO2t0jWMUW")
