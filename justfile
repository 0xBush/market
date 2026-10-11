set dotenv-load := true

sqlc := "go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1"
goose := "go run github.com/pressly/goose/v3/cmd/goose@v3.28.0"
swagger := "go run github.com/swaggo/swag/cmd/swag@v1.16.6"

default:
    @just --list

up:
    docker compose up -d --wait

down:
    docker compose down

cleanup:
    docker compose down -v
    rm -rf build

setup:
    [ -f .env ] || cp .env.example .env
    just up

run:
    go run ./cmd/api

migrate:
    go run ./cmd/migrate

migrate-down:
    go run ./cmd/migrate -down

migrate-new name:
    {{goose}} -dir migrations create {{name}} sql

sqlc:
    {{sqlc}} generate

swagger:
    {{swagger}} init -g cmd/api/main.go -o docs/openapi --outputTypes go,yaml --parseInternal

build:
    mkdir -p build
    go build -o build ./cmd/api
    go build -o build ./cmd/migrate
