package order

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	defaultListLimit int32 = 50
	maxListLimit     int32 = 100
)

type repo interface {
	Create(ctx context.Context, name string, quantity int32) (Order, error)
	Get(ctx context.Context, id int64) (Order, error)
	List(ctx context.Context, limit, offset int32) ([]Order, error)
}

type Service struct {
	repo repo
}

func NewService(r repo) *Service {
	return &Service{repo: r}
}

func (s *Service) Create(ctx context.Context, in CreateInput) (Order, error) {
	name, err := validateName(in.Name)
	if err != nil {
		return Order{}, err
	}
	if in.Quantity < 1 {
		return Order{}, fmt.Errorf("%w: quantity must be at least 1", ErrInvalid)
	}
	return s.repo.Create(ctx, name, in.Quantity)
}

func (s *Service) Get(ctx context.Context, id int64) (Order, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) List(ctx context.Context, limit, offset int32) ([]Order, error) {
	if limit == 0 {
		limit = defaultListLimit
	}
	if limit < 1 || limit > maxListLimit {
		return nil, fmt.Errorf("%w: limit must be 1-%d", ErrInvalid, maxListLimit)
	}
	if offset < 0 {
		return nil, fmt.Errorf("%w: offset must be >= 0", ErrInvalid)
	}
	return s.repo.List(ctx, limit, offset)
}

func validateName(s string) (string, error) {
	s = strings.TrimSpace(s)
	n := utf8.RuneCountInString(s)
	if n < 1 || n > 100 {
		return "", fmt.Errorf("%w: name must be 1-100 characters", ErrInvalid)
	}
	return s, nil
}
