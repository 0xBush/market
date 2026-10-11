package auth

import (
	"context"
	"encoding/hex"
	"errors"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type Repository struct {
	rdb *redis.Client
}

func NewRepository(rdb *redis.Client) *Repository {
	return &Repository{rdb: rdb}
}

func refreshKey(tokenHash []byte) string {
	return "refresh:" + hex.EncodeToString(tokenHash)
}

func (r *Repository) CreateRefresh(ctx context.Context, userID int64, tokenHash []byte, expiresAt time.Time) error {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return ErrUnauthorized
	}
	return r.rdb.Set(ctx, refreshKey(tokenHash), userID, ttl).Err()
}

func (r *Repository) DeleteRefresh(ctx context.Context, tokenHash []byte) error {
	return r.rdb.Del(ctx, refreshKey(tokenHash)).Err()
}

func (r *Repository) RotateRefresh(ctx context.Context, oldHash, newHash []byte, expiresAt time.Time) (int64, error) {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return 0, ErrUnauthorized
	}

	val, err := r.rdb.GetDel(ctx, refreshKey(oldHash)).Result()
	if errors.Is(err, redis.Nil) {
		return 0, ErrUnauthorized
	}
	if err != nil {
		return 0, err
	}

	userID, err := strconv.ParseInt(val, 10, 64)
	if err != nil || userID < 1 {
		return 0, ErrUnauthorized
	}

	if err := r.rdb.Set(ctx, refreshKey(newHash), userID, ttl).Err(); err != nil {
		return 0, err
	}
	return userID, nil
}
