-- name: CreateOrder :one
INSERT INTO orders (name, quantity)
VALUES (sqlc.arg(name), sqlc.arg(quantity))
RETURNING id, name, quantity, created_at, updated_at;

-- name: GetOrder :one
SELECT id, name, quantity, created_at, updated_at
FROM orders
WHERE id = sqlc.arg(id);

-- name: ListOrders :many
SELECT id, name, quantity, created_at, updated_at
FROM orders
ORDER BY id;
