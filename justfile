set dotenv-load := true

sqlc := "go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1"
goose := "go run github.com/pressly/goose/v3/cmd/goose@v3.28.0"

default:
    @just --list

up:
    docker compose up -d --wait

down:
    docker compose down

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

build:
    mkdir -p build
    go build -o build ./cmd/api
    go build -o build ./cmd/migrate
