package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func MigrateUp(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	p, db, err := newProvider(pool, dir)
	if err != nil {
		return err
	}
	defer db.Close()
	if _, err := p.Up(ctx); err != nil {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}

func MigrateDown(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	p, db, err := newProvider(pool, dir)
	if err != nil {
		return err
	}
	defer db.Close()
	if _, err := p.Down(ctx); err != nil {
		return fmt.Errorf("migrate down: %w", err)
	}
	return nil
}

func newProvider(pool *pgxpool.Pool, dir string) (*goose.Provider, *sql.DB, error) {
	db := stdlib.OpenDBFromPool(pool)
	p, err := goose.NewProvider(goose.DialectPostgres, db, os.DirFS(dir))
	if err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("goose provider: %w", err)
	}
	return p, db, nil
}
