# market

Go REST API for a marketplace. The first resource is **orders**.

The HTTP API is small. The layout is the pattern for the next feature: one package under `internal/`, sqlc queries, goose migrations, chi.

## Stack

- chi router (`net/http` underneath)
- `go-chi/httplog` for slog access logs and panic recover
- Postgres 18 with `pgx/v5`
- sqlc for typed queries
- goose for migrations
- `log/slog`

## Layout

```
cmd/api          HTTP server
cmd/migrate      goose up / down
internal/config  env vars
internal/database
  database.go    pgx pool
  migrate.go     goose runner
  sqlc/          generated code (committed)
internal/httputil
internal/server  chi router, shutdown
internal/order   domain, repository, service, HTTP
migrations/      schema (sqlc reads this too)
```

An order is `id`, `name`, `quantity`, `created_at`, `updated_at`.

A request hits the handler, the handler calls the service, the service calls the repository, the repository calls sqlc. Interfaces sit with the caller.

Each module registers its own routes on a `chi.Router` (`Handler.Register`) and maps its own errors to HTTP. `internal/server` mounts `/healthz`, `/readyz`, chi middleware, and the modules in `routes.go`. `cmd/api` is process bootstrap only.

## Run locally

Needs Docker, Go 1.26, and [just](https://github.com/casey/just).

```sh
cp .env.example .env
just up
just migrate
just run
```

Two binaries:

```sh
go run ./cmd/api       # HTTP on :8080
go run ./cmd/migrate   # apply migrations, then exit
go run ./cmd/migrate -down
```

New migration:

```sh
just migrate-new add_something
```

After you change `internal/order/queries.sql` or a migration:

```sh
just sqlc
```

## API

| Method | Path | Action |
|---|---|---|
| GET | `/healthz` | process is up |
| GET | `/readyz` | Postgres answers ping |
| POST | `/orders` | create |
| GET | `/orders` | list |
| GET | `/orders/{id}` | get one |

```sh
curl -s localhost:8080/healthz

curl -s -X POST localhost:8080/orders \
  -H 'Content-Type: application/json' \
  -d '{"name":"widget","quantity":3}'

curl -s localhost:8080/orders
curl -s localhost:8080/orders/1
```
