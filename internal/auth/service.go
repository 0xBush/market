package auth

import (
	"context"
	"errors"
	"time"

	"github.com/0xbush/market/internal/config"
	"github.com/0xbush/market/internal/user"
)

type accounts interface {
	Register(ctx context.Context, in user.Credentials) (user.User, error)
	Authenticate(ctx context.Context, in user.Credentials) (user.User, error)
	GetByID(ctx context.Context, id int64) (user.User, error)
}

type repo interface {
	CreateRefresh(ctx context.Context, userID int64, tokenHash []byte, expiresAt time.Time) error
	DeleteRefresh(ctx context.Context, tokenHash []byte) error
	RotateRefresh(ctx context.Context, oldHash, newHash []byte, expiresAt time.Time) (int64, error)
}

type Service struct {
	accounts accounts
	repo     repo
	cfg      config.Auth
}

func NewService(accounts accounts, r repo, cfg config.Auth) *Service {
	return &Service{accounts: accounts, repo: r, cfg: cfg}
}

func (s *Service) Register(ctx context.Context, in user.Credentials) (user.User, TokenPair, error) {
	u, err := s.accounts.Register(ctx, in)
	if err != nil {
		return user.User{}, TokenPair{}, err
	}
	pair, err := s.issueTokens(ctx, u.ID)
	if err != nil {
		return user.User{}, TokenPair{}, err
	}
	return u, pair, nil
}

func (s *Service) Login(ctx context.Context, in user.Credentials) (user.User, TokenPair, error) {
	u, err := s.accounts.Authenticate(ctx, in)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return user.User{}, TokenPair{}, ErrUnauthorized
		}
		return user.User{}, TokenPair{}, err
	}
	pair, err := s.issueTokens(ctx, u.ID)
	if err != nil {
		return user.User{}, TokenPair{}, err
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

	userID, err := s.repo.RotateRefresh(ctx, hashRefreshToken(rawRefresh), hash, time.Now().Add(s.cfg.RefreshTTL))
	if err != nil {
		return TokenPair{}, err
	}

	access, err := signAccessToken(userID, s.cfg.JWTSecret, s.cfg.AccessTTL)
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

func (s *Service) Me(ctx context.Context, userID int64) (user.User, error) {
	u, err := s.accounts.GetByID(ctx, userID)
	if errors.Is(err, user.ErrNotFound) {
		return user.User{}, ErrUnauthorized
	}
	return u, err
}

func (s *Service) issueTokens(ctx context.Context, userID int64) (TokenPair, error) {
	access, err := signAccessToken(userID, s.cfg.JWTSecret, s.cfg.AccessTTL)
	if err != nil {
		return TokenPair{}, err
	}
	raw, hash, err := newRefreshToken()
	if err != nil {
		return TokenPair{}, err
	}
	if err := s.repo.CreateRefresh(ctx, userID, hash, time.Now().Add(s.cfg.RefreshTTL)); err != nil {
		return TokenPair{}, err
	}
	return TokenPair{Access: access, Refresh: raw}, nil
}
