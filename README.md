# Calculator

A full-stack calculator: a Go API that evaluates arithmetic expressions and a React (Vite + HeroUI v3) frontend that drives it. The frontend is a thin client — all parsing, validation, and arithmetic happens server-side.

## Quick start

```bash
docker compose up --build
# → http://localhost:3000  (the web UI)
# → http://localhost:8080  (the API, exposed for direct curl-ing)
```

## Local development

Run each side independently without Docker.

**Backend:**
```bash
cd api
go run ./cmd/server      # listens on :8080
```

**Frontend:**
```bash
cd web
npm install
npm run dev              # http://localhost:5173 (proxies /api/* → :8080)
```

## API reference

`POST /api/calculate`

**Request:**
```json
{ "expression": "(1+1)*4" }
```

**Success (HTTP 200):**
```json
{ "result": "8" }
```

**Error (HTTP 4xx/5xx):**
```json
{ "error": "division by zero" }
```

Status codes:

| Code | Meaning                                      |
|------|----------------------------------------------|
| 200  | Success                                       |
| 400  | Empty, malformed, or unparseable expression  |
| 415  | Wrong `Content-Type`                          |
| 422  | Division by zero / mathematically undefined  |
| 500  | Internal server error                         |

**Examples:**
```bash
# Success
curl -s -X POST http://localhost:8080/api/calculate \
  -H 'Content-Type: application/json' \
  -d '{"expression":"(1+1)*4"}'
# {"result":"8"}

# Percent-of — "50% of 100"
curl -s -X POST http://localhost:8080/api/calculate \
  -H 'Content-Type: application/json' \
  -d '{"expression":"50 % 100"}'
# {"result":"50"}

# Division by zero
curl -s -X POST http://localhost:8080/api/calculate \
  -H 'Content-Type: application/json' \
  -d '{"expression":"1/0"}'
# {"error":"division by zero"}

# Invalid expression
curl -s -X POST http://localhost:8080/api/calculate \
  -H 'Content-Type: application/json' \
  -d '{"expression":"1+"}'
# {"error":"invalid expression"}
```

Also: `GET /healthz` → `{"status":"ok"}` for liveness checks.

## Supported operations

| Operator        | Meaning                            | Example          |
|-----------------|------------------------------------|------------------|
| `+ - * /`       | Basic arithmetic                   | `1+2*3` → `7`    |
| `^`             | Exponentiation                     | `2^10` → `1024`  |
| `sqrt(x)`       | Square root                        | `sqrt(16)` → `4` |
| `%`             | **Percent-of**: `a % b = (a/100)*b`| `50 % 100` → `50`|
| `(` `)`         | Grouping                           | `(1+2)*3` → `9`  |

**Precedence** (highest → lowest): parentheses → function calls → `^` → `* / %` → `+ -`.

**Percent-of is binary only.** `50 %` by itself is rejected. `%` specifically means "X percent of Y", not modulo. This is deliberate — calculators conventionally treat `%` this way and modulo is rarely needed in a UI meant for arithmetic.

## Design decisions

- **`expr-lang/expr` over a hand-written parser.** The library is mature, actively maintained, and covers the grammar we need (arithmetic, functions, parens, precedence) out of the box. A hand-rolled parser would have consumed the whole time budget and delivered nothing novel; letting `expr` do it lets the effort go into correctness tests and UX polish.

- **Percent-of is preprocessed, not custom-operatored.** `expr` treats `%` as modulo and doesn't offer a way to override that cleanly. A regex-driven rewrite before evaluation (`a % b` → `((a)/100)*b`) keeps the calc grammar small and the `expr` integration pristine. The preprocessor also wraps the rewrite in outer parens so it composes correctly with higher-precedence operators (e.g., `2^3 % 100` doesn't swallow the `**`).

- **Integer literals are promoted to floats before evaluation.** This gives IEEE-754 semantics everywhere: `1/0` produces `+Inf` instead of panicking, so the handler can detect division-by-zero uniformly via `math.IsInf`/`math.IsNaN` on the result. No special-casing per operator.

- **Results returned as strings.** The backend owns presentation. Using JSON numbers would force the client to decide how to render `8.000000001` vs `8`, or to deal with precision loss on very large values. A string lets the server answer "what should the user see" authoritatively — trimmed zeros, no trailing `.0` on integer results, scientific notation only for values outside `[1e-6, 1e15]`.

- **All logic on the backend.** The frontend never inspects, parses, or arithmetics. The only UX-level guard is blocking submission of an empty string — and the API rejects empty strings too (defense in depth, not logic duplication). One source of truth = one place to test evaluation behavior.

- **Zustand + TanStack Query, not one or the other.** They don't overlap. Zustand holds UI state (current expression, result, history — things only the browser cares about). TanStack Query handles server state (the in-flight calculate mutation, its error state, retries). Collapsing both into one library would mean either building a cache framework in Zustand or misusing Query as a general state store.

- **Client-persisted history.** The backend is stateless: no DB, no session, no auth, no deployment complexity. History is purely a UX convenience and lives in `localStorage` under a versioned key (`calculator:history:v1`) so we can evolve the schema without colliding with old data. The current expression and result are in-memory only — refreshing the page resets the session but preserves the list.

- **Known limitations.** No auth or rate limiting (API is open). No server-side persistence. No scientific functions beyond `sqrt`. No unit conversions. No BigInt / arbitrary precision — everything is `float64`.

- **Equation doesn't clear on resolve.** This is by design, for when the user wants to do consecutive similar calculations.

## Testing

**API:**
```bash
cd api
go test ./...
go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out
```
Current coverage: `internal/calc` ~93%, `internal/httpapi` ~96%.

**Web:**
```bash
cd web
npm test                        # run once
npm run test:watch              # watch mode
npm run test:coverage           # with coverage report
```
31 tests covering stores, the API wrapper, and the main calculator + history components.

## Project structure

```
.
├── api/                      # Go backend
│   ├── cmd/server/           # HTTP entry point
│   ├── internal/calc/        # preprocess, evaluate, format
│   ├── internal/httpapi/     # handler, error mapping, middleware
│   └── Dockerfile            # multi-stage → distroless static
├── web/                      # React frontend
│   ├── src/
│   │   ├── api/              # fetch wrapper
│   │   ├── hooks/            # TanStack Query mutation
│   │   ├── stores/           # Zustand: calc, history, theme
│   │   ├── components/       # Calculator, Display, Keypad, History
│   │   └── styles/           # Tailwind v4 + HeroUI v3 CSS
│   ├── Dockerfile            # multi-stage → nginx:alpine
│   └── nginx.conf            # SPA + /api/* reverse proxy
└── docker-compose.yml        # ties them together
```
