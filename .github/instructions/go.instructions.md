---
description: "Canonical Go coding rules for the Recipe App backend. Referenced by all agents and skills. Covers project layout, chi handlers, database/sql repositories, error handling, validation, auth, and testing for the actual stack (chi + raw SQL, no ORM)."
applyTo: "backend/**/*.go"
---

# Go Coding Rules — Recipe App Backend

These are the canonical rules for writing and reviewing Go code in `backend/`. They reflect
the stack as it is actually implemented: **chi v5 + `database/sql` (raw parameterized SQL) +
JWT/bcrypt, no ORM**. Load the shared [Go rule pack](../references/go-shared-rules.md) for
the condensed checklist that agents and skills share.

## 1. Project Layout

```
backend/
  cmd/                 # main.go: entry point, DB open, table creation, route wiring
  internal/
    appmiddleware/     # chi middleware: auth, CORS, rate limit, error, security headers
    database/          # embedded schema.sql + ApplySchema()
    handlers/          # HTTP handlers (one struct per area: APIHandler, AuthHandler, ...)
    logger/            # structured logger, context helpers (logger.FromContext)
    models/            # domain structs + Validate() methods
    repositories/      # data access: database/sql, one struct per aggregate
    services/          # business logic (currently empty — add only when justified)
    storage/           # storage helpers (currently empty)
  web/                 # templates + static assets served by the Go process
```

- Module path is `recipe-app`; import internal packages as `recipe-app/internal/...`.
- Keep packages cohesive. Do not create a package until there is real code for it.

## 2. Naming & Style

- Run `gofmt`/`go vet`; code must be gofmt-clean (tabs, standard import grouping).
- Exported identifiers use `PascalCase`, unexported use `camelCase`. Acronyms stay
  uppercase (`ID`, `URL`, `JWT`, `HTTP`).
- Constructors are `NewXxx(...) *Xxx` and inject dependencies (see `NewRecipeRepository`,
  `NewAPIHandler`). No global mutable state; pass dependencies through constructors.
- Group imports in three blocks: standard library, third-party, then `recipe-app/...`.

## 3. HTTP Handlers (chi)

- Handler methods have signature `func (h *Handler) Name(w http.ResponseWriter, r *http.Request)`.
- Read path params with `chi.URLParam(r, "id")` and query params with `r.URL.Query().Get(...)`.
- Decode JSON bodies with `json.NewDecoder(r.Body).Decode(&v)`; on error return
  `http.StatusBadRequest`.
- Always call the model's `Validate()` before persisting; map validation failures to `400`.
- Set `Content-Type` before writing the body. JSON responses use
  `w.Header().Set("Content-Type", "application/json")` then `json.NewEncoder(w).Encode(...)`.
- Support the existing HTMX pattern: when `r.Header.Get("HX-Request") == "true"`, render the
  matching `html/template` partial; otherwise return JSON.
- **Status codes**: `200` ok, `201` created, `204` no content, `400` validation, `401`
  unauthorized, `403` forbidden, `404` not found, `409` conflict, `500` internal.
- Keep handlers thin: parse → validate → call repository (or service) → write response. No
  SQL in handlers.

### Known debt to respect, not copy
- Some handlers pass a `ctx interface{}` parameter and call `r.Context()` internally. Do not
  propagate this pattern in new code — thread the real `context.Context` through repository
  calls instead. Do not silently "fix" existing signatures unless the task asks for it.

## 4. Repositories (`database/sql`)

- One repository struct per aggregate, holding `db *sql.DB`.
- **Always** use placeholders and pass args separately. Never concatenate user input into
  SQL. SQLite uses `?` placeholders.
- For dynamic filters, append ` AND col = ?` to the query and append the matching value to
  the args slice (existing `GetRecipes` pattern) — the values still go through placeholders,
  never string interpolation of user data.
- Single-row reads use `QueryRow(...).Scan(...)`; translate `sql.ErrNoRows` into a
  not-found error instead of returning it raw.
- Multi-row reads use `Query(...)`, **always** `defer rows.Close()`, scan in the loop, and
  check `rows.Err()` after the loop.
- Wrap every returned error with context: `fmt.Errorf("failed to get recipe: %w", err)`.
- Multi-step writes that must be atomic should use `db.BeginTx`/`tx.Commit`/`tx.Rollback`.
  (Today `CreateRecipe` writes children without a transaction — when touching this code,
  prefer wrapping the related writes in a transaction.)

## 5. Models & Validation

- Domain structs live in `internal/models` with `json:"..."` and `db:"..."` tags.
- Validation is explicit: a `func (x *X) Validate() error` returning the first failure via
  `fmt.Errorf`. No struct-tag validation library.
- Enumerated fields (e.g. `difficulty` ∈ {easy, medium, hard}) are validated in `Validate()`
  and mirrored by a `CHECK` constraint in the schema.

## 6. Auth & Security

- JWT via `golang-jwt/jwt/v5`; passwords hashed with `bcrypt`. Never log or return hashes.
- Read the JWT secret from the `JWT_SECRET` env var; never hardcode a production secret.
- Protect routes with the existing middleware (`authService.AuthMiddleware`,
  `OptionalAuthMiddleware`) wired in `cmd/main.go`.
- Keep the security middleware chain intact (RequestID, Recoverer, RequestLogger,
  ErrorHandler, CORS, RateLimit, SecurityHeaders).

## 7. Error Handling

- Check every error. Wrap with `%w` and a message describing the failing operation.
- Do not return database/internal detail to clients on untrusted paths; log it with the
  context logger (`logger.FromContext(r.Context())`) and return a generic message.
- No empty error branches; never swallow an error silently.

## 8. Logging

- Use the project logger (`internal/logger`). Prefer the request-scoped logger via
  `logger.FromContext(r.Context())`. Use structured key/value pairs
  (`log.Info("Creating new recipe")`, `log.Error("...", "error", err)`).

## 9. Testing

- Table-driven tests with subtests (`t.Run(tt.name, ...)`).
- Handlers: use `net/http/httptest` (`httptest.NewRequest`, `httptest.NewRecorder`) as in
  `internal/handlers/api_test.go`. For chi path params, inject a `chi.RouteContext` when
  needed.
- Test the behavior, not the implementation. Cover the happy path plus validation and
  error branches.
- A change is not done until `go build ./...` and `go test ./...` pass from `backend/`.

## Prohibited

- ❌ ORMs or query builders (the codebase uses raw `database/sql`).
- ❌ String-concatenated SQL with user input.
- ❌ New dependencies without explicit justification + `go mod tidy`.
- ❌ Ignoring errors with `_` (except deliberately, with a comment).
- ❌ Global mutable state; use constructor injection.
- ❌ Returning raw internal errors to clients on untrusted input paths.
