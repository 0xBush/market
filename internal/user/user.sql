-- name: CreateUser :one
INSERT INTO users (email, password_hash)
VALUES (sqlc.arg(email), sqlc.arg(password_hash))
RETURNING id, email, created_at;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, created_at
FROM users
WHERE email = sqlc.arg(email);

-- name: GetUserByID :one
SELECT id, email, created_at
FROM users
WHERE id = sqlc.arg(id);

-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
VALUES (sqlc.arg(user_id), sqlc.arg(token_hash), sqlc.arg(expires_at));

-- name: GetRefreshToken :one
SELECT id, user_id, token_hash, expires_at, created_at
FROM refresh_tokens
WHERE token_hash = sqlc.arg(token_hash);

-- name: DeleteRefreshToken :execrows
DELETE FROM refresh_tokens
WHERE token_hash = sqlc.arg(token_hash);
