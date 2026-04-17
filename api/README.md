# Calculator API

A small Go HTTP service that evaluates arithmetic expressions for the calculator UI in `/web`.

## Run

```bash
go run ./cmd/server          # listens on :8080 by default
ADDR=:9000 go run ./cmd/server
```

## Test

```bash
go test ./...
go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out
```

Current coverage: `internal/calc` ~93%, `internal/httpapi` ~96%.

## Endpoint

`POST /api/calculate` — body `{"expression": "<string>"}`, returns `{"result": "<string>"}` on 200 or `{"error": "<message>"}` on 4xx/5xx.

`GET /healthz` — liveness probe, returns `{"status": "ok"}`.

## Layout

- `cmd/server/main.go` — wiring only
- `internal/calc/` — preprocessing (`^` → `**`, percent-of rewrite), evaluation via `expr-lang/expr`, result formatting, typed sentinel errors
- `internal/httpapi/` — handler, error → status mapping, CORS + request logging middleware
