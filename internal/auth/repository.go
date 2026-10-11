package auth

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xbush/market/internal/database/sqlc"
)

type Repository struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
}

func NewRepository(pool *pgxpool.Pool, q *sqlc.Queries) *Repository {
	return &Repository{pool: pool, q: q}
}

func (r *Repository) CreateRefresh(ctx context.Context, userID int64, tokenHash []byte, expiresAt time.Time) error {
	return r.q.CreateRefreshToken(ctx, userID, tokenHash, expiresAt)
}

func (r *Repository) DeleteRefresh(ctx context.Context, tokenHash []byte) error {
	_, err := r.q.DeleteRefreshToken(ctx, tokenHash)
	return err
}

func (r *Repository) RotateRefresh(ctx context.Context, oldHash, newHash []byte, expiresAt time.Time) (int64, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	q := r.q.WithTx(tx)

	row, err := q.GetRefreshToken(ctx, oldHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrUnauthorized
	}
	if err != nil {
		return 0, err
	}
	if !time.Now().Before(row.ExpiresAt) {
		_, _ = q.DeleteRefreshToken(ctx, oldHash)
		_ = tx.Commit(ctx)
		return 0, ErrUnauthorized
	}

	n, err := q.DeleteRefreshToken(ctx, oldHash)
	if err != nil {
		return 0, err
	}
	if n == 0 {
		return 0, ErrUnauthorized
	}

	if err := q.CreateRefreshToken(ctx, row.UserID, newHash, expiresAt); err != nil {
		return 0, err
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return row.UserID, nil
}
