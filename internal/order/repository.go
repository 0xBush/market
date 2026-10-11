package order

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/0xbush/market/internal/database/sqlc"
)

type Repository struct {
	q *sqlc.Queries
}

func NewRepository(q *sqlc.Queries) *Repository {
	return &Repository{q: q}
}

func (r *Repository) Create(ctx context.Context, userID int64, name string, quantity int32) (Order, error) {
	row, err := r.q.CreateOrder(ctx, name, quantity, userID)
	if err != nil {
		return Order{}, err
	}
	return toOrder(row), nil
}

func (r *Repository) Get(ctx context.Context, id, userID int64) (Order, error) {
	row, err := r.q.GetOrder(ctx, id, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Order{}, ErrNotFound
	}
	if err != nil {
		return Order{}, err
	}
	return toOrder(row), nil
}

func (r *Repository) List(ctx context.Context, userID int64) ([]Order, error) {
	rows, err := r.q.ListOrders(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]Order, 0, len(rows))
	for _, row := range rows {
		out = append(out, toOrder(row))
	}
	return out, nil
}

func toOrder(row sqlc.Order) Order {
	return Order{
		ID:        row.ID,
		UserID:    row.UserID,
		Name:      row.Name,
		Quantity:  row.Quantity,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
