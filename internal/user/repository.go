package user

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/0xbush/market/internal/database/sqlc"
)

type Repository struct {
	q *sqlc.Queries
}

func NewRepository(q *sqlc.Queries) *Repository {
	return &Repository{q: q}
}

func (r *Repository) Create(ctx context.Context, email, passwordHash string) (User, error) {
	row, err := r.q.CreateUser(ctx, email, passwordHash)
	if isUniqueViolation(err) {
		return User{}, ErrConflict
	}
	if err != nil {
		return User{}, err
	}
	return User{ID: row.ID, Email: row.Email, CreatedAt: row.CreatedAt}, nil
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (User, string, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, "", ErrNotFound
	}
	if err != nil {
		return User{}, "", err
	}
	return User{ID: row.ID, Email: row.Email, CreatedAt: row.CreatedAt}, row.PasswordHash, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (User, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, err
	}
	return User{ID: row.ID, Email: row.Email, CreatedAt: row.CreatedAt}, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
