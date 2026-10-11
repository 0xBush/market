# market

Go REST API for a marketplace. The first resource is **orders**. Accounts use email and password; access is a short JWT cookie, refresh is an opaque cookie stored hashed in Postgres.

The HTTP API is small. A domain resource is one package under `internal/` with sqlc queries, a service, and HTTP.

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
internal/user    accounts, JWT, refresh cookies
internal/order   domain, repository, service, HTTP
migrations/      schema (sqlc reads this too)
docs/            generated Swagger (committed; `just swagger`)
```

A user is `id`, `email`, `created_at`. An order is `id`, `user_id`, `name`, `quantity`, `created_at`, `updated_at`.

A request hits the handler, the handler calls the service, the service calls the repository, the repository calls sqlc. Interfaces sit with the caller.

`internal/server` mounts `/healthz`, `/readyz`, auth, and orders. `cmd/api` is process bootstrap only.

## Run locally

Needs Docker, Go 1.26, and [just](https://casey.github.io/just/).

```sh
just setup
just run
```

`just setup` copies `.env` if missing, starts Postgres, and applies migrations.

Stop the stack and remove `build/`:

```sh
just cleanup
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

After you change `internal/*/*.sql` or a migration:

```sh
just sqlc
```

After you change swag comments on handlers:

```sh
just swagger
```

That regenerates `docs/`. Serve the UI at `/swagger/index.html` while the API is running.

## API

| Method | Path | Action |
|---|---|---|
| GET | `/healthz` | process is up |
| GET | `/readyz` | Postgres answers ping |
| GET | `/swagger/index.html` | Swagger UI |
| POST | `/register` | create account, set cookies |
| POST | `/login` | set cookies |
| POST | `/refresh` | rotate refresh, new access cookie |
| POST | `/logout` | revoke refresh, clear cookies |
| GET | `/me` | current user (access cookie) |
| POST | `/orders` | create own order |
| GET | `/orders` | list own orders |
| GET | `/orders/{id}` | get own order |

Cookies: `access_token` (JWT, 15m) and `refresh_token` (opaque, 7d). Both HttpOnly, `SameSite=Lax`. Orders and `/me` require a valid access cookie; otherwise 401.

```sh
curl -s localhost:8080/healthz

curl -s -c cookies -b cookies -X POST localhost:8080/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"ada@example.com","password":"secret123"}'

curl -s -c cookies -b cookies localhost:8080/me

curl -s -c cookies -b cookies -X POST localhost:8080/orders \
  -H 'Content-Type: application/json' \
  -d '{"name":"widget","quantity":3}'

curl -s -c cookies -b cookies localhost:8080/orders
curl -s -c cookies -b cookies localhost:8080/orders/1

# when access JWT has expired
curl -s -c cookies -b cookies -X POST localhost:8080/refresh

curl -s -c cookies -b cookies -X POST localhost:8080/logout
```
