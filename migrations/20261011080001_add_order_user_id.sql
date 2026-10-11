-- +goose Up
ALTER TABLE orders
    ADD COLUMN user_id BIGINT NOT NULL REFERENCES users (id);

CREATE INDEX orders_user_id_idx ON orders (user_id);

-- +goose Down
DROP INDEX orders_user_id_idx;

ALTER TABLE orders
    DROP COLUMN user_id;
