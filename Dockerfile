FROM golang:1.26-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/api ./cmd/api
RUN CGO_ENABLED=0 go build -o /out/migrate ./cmd/migrate

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/api /api
COPY --from=build /out/migrate /migrate
COPY migrations /app/migrations
ENV MIGRATIONS_DIR=/app/migrations
USER nonroot
ENTRYPOINT ["/api"]
