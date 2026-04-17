# Calculator — Full-Stack Take-Home

A calculator application with a Go API that evaluates mathematical expressions and a React frontend that collects input and displays results. This is a take-home assignment; the priority is clean, idiomatic code and meaningful tests, not feature breadth.

## Repository Layout

Monorepo with two top-level directories:

```
/
├── api/                 # Go backend
├── web/                 # React frontend
├── docker-compose.yml   # Runs both services together
└── README.md            # Setup, API examples, design decisions
```

Each sub-project has its own `Dockerfile`. The root `docker-compose.yml` ties them together for self-hosted deployment.

## Core Design Principle

**All parsing, validation, and arithmetic happens on the API.** The frontend is a dumb terminal: it collects keystrokes, sends a string to the API, and renders whatever comes back. The frontend never inspects the expression, never does arithmetic, never validates syntax.

The only UX-level guard on the frontend is blocking submission of an empty string (and the API must also reject empty strings — defense in depth, not logic duplication).

---

## Backend (`/api`)

### Stack

- Go 1.22+
- `net/http` standard library router (Go 1.22's pattern-matching router is sufficient for one endpoint)
- [`github.com/expr-lang/expr`](https://github.com/expr-lang/expr) for expression evaluation
- `log/slog` for structured logging
- `golangci-lint` for linting

No web framework. No ORM. No config library. Stdlib plus `expr`.

### Endpoint

**`POST /api/calculate`**

Request:
```json
{ "expression": "(1+1)*4" }
```

Success response (HTTP 200):
```json
{ "result": "8" }
```

Error response (HTTP 400 for malformed input, 422 for semantic errors like division by zero):
```json
{ "error": "division by zero" }
```

The `error` field contains a message ready to display to the user — no error codes, no translation keys. The frontend pipes this straight into a toast.

**Result formatting:** results are returned as strings so the backend controls presentation (e.g., `"8"` not `"8.0000000001"`, trim trailing zeros on floats, avoid scientific notation for reasonable ranges). Format rules:
- Integer results: no decimal point (`"8"`, not `"8.0"`)
- Float results: up to 10 significant digits, trailing zeros trimmed
- Results outside `[1e-6, 1e15]` in magnitude: scientific notation is acceptable

### Supported Operations

- `+`, `-`, `*`, `/`
- `^` for exponentiation
- `sqrt(x)` for square root
- `%` for **percent-of**: `a % b` means "a percent of b", i.e., `(a/100)*b`. So `50 % 100 = 50`.
- Parentheses for grouping

### Percent-of Implementation

`expr-lang/expr` treats `%` as modulo. We preprocess the expression before evaluation:

1. A regex rewrites `<number-or-parenthesized-expr> % <number-or-parenthesized-expr>` to `((<left>)/100)*(<right>)`.
2. The rewritten expression is passed to `expr.Eval`.
3. Single-operand percent (`50%` by itself) is **not supported** — the regex only matches the binary form, and anything else falls through to `expr`, which returns an error. That error surfaces to the user as "invalid expression."

Put the preprocessing in its own package (e.g., `internal/preprocess`) so it's unit-testable in isolation from the evaluator.

### Package Structure

```
api/
├── cmd/server/main.go           # Entry point, wiring only
├── internal/
│   ├── calc/
│   │   ├── calc.go              # Calculate(expression string) (string, error)
│   │   ├── calc_test.go
│   │   ├── preprocess.go        # Percent-of rewrite
│   │   ├── preprocess_test.go
│   │   ├── format.go            # Result formatting
│   │   └── format_test.go
│   └── httpapi/
│       ├── handler.go           # HTTP handler, maps errors to status codes
│       ├── handler_test.go
│       ├── errors.go            # Typed errors: ErrInvalidExpression, ErrDivisionByZero, etc.
│       └── middleware.go        # CORS, request logging
├── go.mod
├── Dockerfile
└── README.md
```

The HTTP handler is thin. It decodes JSON, calls `calc.Calculate`, maps the returned error type to an HTTP status code, and encodes the response. All actual logic is in `internal/calc` and is trivially testable without spinning up an HTTP server.

### Error Handling

Typed sentinel errors in `internal/calc` (or wrapped errors with `errors.Is` support):

- `ErrEmptyExpression` → 400
- `ErrInvalidExpression` → 400 (parse failures from `expr`)
- `ErrDivisionByZero` → 422
- `ErrNonNumericResult` → 422 (e.g., if `expr` somehow returns a non-number)
- Anything else → 500 with a generic message (don't leak internals)

The handler has one `switch` or `errors.Is` chain mapping these to statuses. Error messages in responses are user-facing strings; log the underlying technical detail via `slog`.

Division by zero needs explicit detection. `expr` returns `+Inf` for `1/0` on floats; check for `math.IsInf` and `math.IsNaN` on the result and convert to `ErrDivisionByZero`.

### Testing

Target: **90%+ coverage** on `internal/calc` and `internal/httpapi`.

- Table-driven tests for `calc.Calculate`: every operator, operator precedence, parentheses, percent-of (several compositions), each error type, whitespace, unicode digits if handled.
- Table-driven tests for `preprocess`: verify rewrite correctness, including nested percents and expressions involving parenthesized sub-expressions on either side of `%`.
- Handler tests using `httptest.NewRecorder`: correct status codes per error type, JSON shape on success and failure, malformed JSON, wrong content-type, empty body, missing field.
- Run coverage: `go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out`. Include a coverage summary in the README.

### Dockerfile

Multi-stage:
1. `golang:1.22-alpine` to build a static binary with `CGO_ENABLED=0`.
2. `gcr.io/distroless/static-debian12` (or `scratch`) for the runtime image.

Expose `:8080`. No shell, no package manager, small image.

---

## Frontend (`/web`)

### Stack

- React 19 + TypeScript
- Vite (not CRA — CRA is deprecated)
- Tailwind CSS v4 (CSS-first config, no `tailwind.config.js`)
- HeroUI v3 (`@heroui/react` + `@heroui/styles`, version 3.0.2+)
- TanStack Query v5 for the calculate mutation
- Zustand for local state (current expression, result, history)
- Vitest + React Testing Library + MSW for tests

### HeroUI v3 Setup Notes

HeroUI v3 differs significantly from v2. Do not use v2 patterns.

- Install: `npm i @heroui/styles @heroui/react tailwind-variants`
- In your main CSS file, order matters:
  ```css
  @import "tailwindcss";
  @import "@heroui/styles";
  ```
- **No `HeroUIProvider` wrapper is needed in v3.** Components work out of the box.
- Components use compound patterns: `<Card.Header>`, `<Card.Content>`, etc. Use dot notation.
- Toast: HeroUI v3 exposes a `toast` function (renamed from v2's `addToast`). See https://heroui.com/docs/react/components/toast. Use it directly for error messages from the calculate mutation.

### UI Behavior

The calculator has two display areas:

1. **Expression input** (top) — editable text showing what the user has typed. Cleared only by `Clear` button or `Esc`.
2. **Result display** (below input) — shows the latest successful result. Persists between calculations; only cleared by `Clear`/`Esc`.

**Button grid:** digits 0-9, operators `+ - * / ^ %`, `sqrt`, `(`, `)`, `.`, `Clear`, `=`.

**Keyboard:**
- Digits, operators, parens, `.` — append to input
- `Enter` — submit (same as `=`)
- `Esc` — clear input *and* result
- `Backspace` — delete last character (bonus, if time permits)

**Submission flow:**
1. User presses `=` or `Enter`.
2. If the expression is empty, do nothing (no request fired).
3. Otherwise, TanStack Query mutation fires `POST /api/calculate`.
4. On success: update the result display, add the `expression → result` pair to history (Zustand). **Do not clear the input.** The user can continue editing from where they left off.
5. On error (HTTP 4xx/5xx or network failure): show a toast with the error message from the response body (or a generic "Network error" if there's no body). Leave input and result as-is.

### State Management

Zustand store(s):

```ts
// Expression + result state
interface CalcState {
  expression: string;
  result: string | null;
  setExpression: (s: string) => void;
  appendToExpression: (s: string) => void;
  setResult: (r: string) => void;
  clear: () => void;
}

// History state
interface HistoryState {
  history: Array<{ expression: string; result: string }>;
  add: (entry: { expression: string; result: string }) => void;
  clear: () => void;
}
```

**History rules:**
- Max 5 entries (FIFO — oldest drops off).
- Do not add if the new entry's `expression` equals the most recent entry's `expression` (dedupe consecutive duplicates only; older duplicates are allowed to resurface).
- History persists across page refreshes via Zustand's `persist` middleware backed by `localStorage`. The key should be namespaced (e.g., `calculator:history:v1`) so we can version the schema later if needed.
- The current `expression` and `result` are **not** persisted — they reset on refresh. Only the history list survives.
- No server-side persistence. The API is stateless; history is purely a UI concern.
- Clicking a history entry replaces the current `expression` with that entry's expression. Result display is not touched (user hasn't submitted yet).

### History UI

A panel (sidebar on desktop, collapsible section on mobile) listing the 5 most recent calculations. Each entry shows the expression followed by `= <result>`. Long expressions truncate with CSS ellipsis to fit the available width — no fixed character cap. The full expression should be available on hover (title attribute) or by clicking.

### TanStack Query Usage

Single mutation: `useCalculateMutation`.

```ts
const mutation = useMutation({
  mutationFn: async (expression: string) => {
    const res = await fetch('/api/calculate', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ expression }),
    });
    const body = await res.json();
    if (!res.ok) {
      throw new Error(body.error ?? 'Calculation failed');
    }
    return body.result as string;
  },
  onSuccess: (result, expression) => {
    setResult(result);
    history.add({ expression, result });
  },
  onError: (err) => {
    showToast(err.message);
  },
});
```

All error handling — thrown parse failures, non-2xx responses, network errors — funnels through `onError` into the toast. The component calling the mutation does not need its own try/catch.

Configure a `QueryClient` at the app root with sensible defaults (no retries for mutations by default — calculation errors aren't transient).

### Responsive Design

Mobile-first. Tailwind breakpoints. On mobile:
- Button grid fills the width
- History panel collapses into a toggleable drawer or an accordion below the calculator

On desktop (≥768px):
- Two-column layout: calculator on the left, history on the right

Keep it simple. No animations beyond what HeroUI provides by default.

### UI Direction

Keep aesthetics minimal. HeroUI's defaults are tasteful — don't fight them. The notes below are the only deliberate deviations.

- **Theme:** light/dark toggle in the UI. Default to `prefers-color-scheme` via `matchMedia`; fall back to light if the system preference is unavailable. Persist the user's explicit choice (if they toggle) in localStorage under `calculator:theme:v1`. If no explicit choice has been made, keep following the system.
- **Accent color:** a warm, muted gold inspired by the HP-12c. Use `#C9A961` as `--color-primary` with a slightly darker shade (`#A88544` or similar) for hover/active states. This is not bright yellow and not orange — if it looks like a school bus or a traffic cone, it's wrong. HeroUI's `primary` token cascades through buttons, focus rings, and accents; set it once in the CSS theme block.
- **Display typography:** monospace with tabular figures for both the expression input and the result. `ui-monospace, "SF Mono", "JetBrains Mono", Menlo, monospace` is fine — no webfont dependency. Digits must align vertically across lines.
- **Button roles:** digits use HeroUI's neutral/default variant; operators use a subtle secondary or tinted variant to distinguish them; `=` is the primary filled button (this is where the gold shows up most prominently); `Clear` is danger-colored but understated (outlined, not filled).
- **Layout:**
  - Desktop (≥768px): two columns. Calculator left, history right.
  - Mobile (<768px): single column, history stacked below the calculator. No drawer, no toggle — just scroll.
- **Density:** tight. Buttons in the keypad should feel tactile and close-packed, not airy. Think calculator, not dashboard.
- **Icons:** avoid. Text labels on every button. If a clear-icon is genuinely needed somewhere, use `lucide-react`.

Everything else — border radii, shadows, transitions, focus rings — take HeroUI's defaults.

### Package Structure

```
web/
├── src/
│   ├── main.tsx
│   ├── App.tsx
│   ├── api/
│   │   └── calculate.ts         # fetch wrapper + types
│   ├── hooks/
│   │   └── useCalculateMutation.ts
│   ├── stores/
│   │   ├── calcStore.ts         # Zustand: expression, result
│   │   └── historyStore.ts      # Zustand: history list
│   ├── components/
│   │   ├── Calculator.tsx       # Container
│   │   ├── Display.tsx          # Expression + result panels
│   │   ├── Keypad.tsx           # Button grid
│   │   └── History.tsx          # Recall panel
│   ├── styles/
│   │   └── globals.css          # @import "tailwindcss"; @import "@heroui/styles";
│   └── test/
│       ├── setup.ts             # Vitest + RTL setup
│       └── handlers.ts          # MSW handlers
├── index.html
├── vite.config.ts
├── tsconfig.json
├── package.json
├── Dockerfile
└── README.md
```

### Testing

Target: meaningful coverage of components, stores, and the mutation hook. Aim for 85%+ but don't chase the number with trivial tests.

- **Store tests:** pure unit tests on `calcStore` and `historyStore` — state transitions, the dedupe rule, the 5-entry cap. For `historyStore`, verify the `persist` middleware rehydrates from localStorage correctly (mock `localStorage` in the test setup).
- **Component tests:** using RTL, cover keypad clicks, keyboard handling (`Enter`, `Esc`), clear behavior, history rendering, clicking a history entry restores the expression.
- **Mutation tests:** with MSW, cover the success path (result displayed, history updated), 4xx error (toast shown with API message), 5xx error (toast shown), network error (toast shown with fallback message).
- **No e2e tests** — out of scope for the time budget.

Coverage: `vitest run --coverage`. Include summary in README.

### Dockerfile

Multi-stage:
1. `node:20-alpine` to install dependencies and run `vite build`.
2. `nginx:alpine` serving `/dist` on port 80, with a minimal `nginx.conf` that proxies `/api/*` to the backend container (so the frontend can call `/api/calculate` without CORS in production).

In Docker Compose, the frontend's nginx proxies to the backend service by name (`http://api:8080`).

---

## Docker Compose

Root `docker-compose.yml`:

- `api` service built from `./api/Dockerfile`, exposed on `:8080` internally, health check on `/healthz` (add a trivial health endpoint).
- `web` service built from `./web/Dockerfile`, exposed on `:3000` externally, depends on `api`.
- One user-facing port: `localhost:3000`. All API calls go through the nginx proxy inside the `web` container.

For local development (hot reload), a separate `docker-compose.dev.yml` is optional — not required for the deliverable. Just document how to run `go run ./cmd/server` and `npm run dev` directly in the README.

---

## README (Top-Level)

The root README must include, in this order:

1. **Overview** — one paragraph on what the project is.
2. **Quick start** — `docker compose up` and you're at `localhost:3000`.
3. **Local development** — how to run API and web separately without Docker.
4. **API reference** — endpoint, request/response shapes, error codes, 3-4 example `curl` calls covering success and errors.
5. **Supported operations** — list with precedence notes and the explicit percent-of definition (`50 % 100 = 50`).
6. **Design decisions** — this section is important. Cover:
   - Why `expr-lang/expr` instead of a hand-written parser (maintained, mature, covers the grammar we need, lets us spend time on testing and UX).
   - Why percent-of is preprocessed rather than implemented as a custom operator in `expr`.
   - Why results are returned as strings.
   - Why all logic is on the backend.
   - Why Zustand + TanStack Query instead of one or the other (Zustand for UI state, TanStack Query for server state — they don't overlap).
   - Why client-side-persisted history (UX convenience, keeps the API stateless — no DB, no session management, no deployment complexity).
   - Known limitations (no auth, no rate limiting, no persistence, no scientific functions beyond sqrt).
7. **Testing** — how to run tests and view coverage for both projects.
8. **Project structure** — brief.

---

## Explicit Constraints

- **Do not** implement expression parsing or arithmetic on the frontend. If you find yourself writing something that looks like validation beyond "is the string empty," stop.
- **Do not** use localStorage, sessionStorage, or IndexedDB for anything other than the history list. Current expression and result are in-memory only.
- **Do not** use HeroUI v2 patterns or APIs. Verify v3 component APIs against the current docs.
- **Do not** introduce additional dependencies beyond what's listed here without documenting why in the README.
- **Do not** add Prettier or ESLint configs that fight with the HeroUI / Tailwind conventions — use sane defaults.
- **Do** keep the frontend and backend independently runnable. Either should start on its own for local dev.
- **Do** make the API return user-ready error messages — no error codes, no i18n keys, no stack traces.
- **Do** keep PRs of this work small and the code boring. This is a take-home; the reviewer wants to see judgment, not cleverness.

## Time Budget Reminder

The assignment specifies 2-4 hours. Budget suggestion:

- Backend (parser integration + preprocess + handler + tests): ~60-75 min
- Frontend (components + stores + mutation + tests): ~75-90 min
- Docker + docker-compose: ~20 min
- README (this is a graded deliverable, not an afterthought): ~20-30 min
- Buffer: remainder

If something is running over, cut scope from the frontend (skip responsive polish, skip Backspace) before cutting tests or the README.
