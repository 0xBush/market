-- name: CreateOrder :one
INSERT INTO orders (name, quantity, user_id)
VALUES (sqlc.arg(name), sqlc.arg(quantity), sqlc.arg(user_id))
RETURNING id, name, quantity, created_at, updated_at, user_id;

-- name: GetOrder :one
SELECT id, name, quantity, created_at, updated_at, user_id
FROM orders
WHERE id = sqlc.arg(id)
  AND user_id = sqlc.arg(user_id);

-- name: ListOrders :many
SELECT id, name, quantity, created_at, updated_at, user_id
FROM orders
WHERE user_id = sqlc.arg(user_id)
ORDER BY id;
