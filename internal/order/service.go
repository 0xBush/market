package order

import (
	"context"
	"strings"
	"unicode/utf8"
)

type repo interface {
	Create(ctx context.Context, name string, quantity int32) (Order, error)
	Get(ctx context.Context, id int64) (Order, error)
	List(ctx context.Context) ([]Order, error)
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
		return Order{}, &ValidationError{Field: "quantity", Msg: "must be at least 1"}
	}
	return s.repo.Create(ctx, name, in.Quantity)
}

func (s *Service) Get(ctx context.Context, id int64) (Order, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]Order, error) {
	return s.repo.List(ctx)
}

func validateName(s string) (string, error) {
	s = strings.TrimSpace(s)
	n := utf8.RuneCountInString(s)
	if n < 1 || n > 100 {
		return "", &ValidationError{Field: "name", Msg: "must be 1-100 characters"}
	}
	return s, nil
}
