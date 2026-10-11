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
