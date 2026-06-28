# Backend Go Agent Instructions

> **Authoritative rules live in [`.github/`](../.github/).** When this file conflicts
> with `.github/copilot-instructions.md` or `.github/instructions/go.instructions.md`,
> **follow `.github/`** — those describe the code as it exists today. This document
> summarizes the same reality for the Go service in `backend/`.

## Tech Stack (as implemented)
- **Language:** Go 1.25
- **Router:** chi v5 (`github.com/go-chi/chi/v5`)
- **Data access:** `database/sql` with raw, parameterized SQL — **no ORM**
- **Driver:** `mattn/go-sqlite3` (SQLite, used by `cmd/main.go` at runtime)
- **Authentication:** JWT (`github.com/golang-jwt/jwt/v5`) + bcrypt
  (`golang.org/x/crypto/bcrypt`)
- **Validation:** no third-party library — models expose `Validate() error` methods
- **IDs:** UUIDs generated in Go (`github.com/google/uuid`) before INSERT
- **Templates:** `html/template` + HTMX for server-rendered HTML

## Project Structure (actual)
```
backend/
  cmd/main.go                  # entry point: wires chi routes -> handlers -> repositories
  internal/
    handlers/                  # HTTP handlers (api, auth, user, collection, web)
    repositories/              # data access (database/sql, raw parameterized SQL)
    models/                    # domain entities + Validate() methods
    appmiddleware/             # JWT auth, CORS, logging, recover, rate limit, security headers
    database/                  # embedded schema.sql + ApplySchema()
    services/                  # currently empty: add only when logic outgrows a handler
    storage/                   # currently empty
  web/                         # html/template views + static assets (HTMX)
```

## Architecture & Layering
- Request flow: **chi route → handler → repository → `database/sql`**.
- **Handlers** are thin: parse/validate input, call repositories, write JSON or HTML.
  Inject dependencies via small interfaces (e.g. `RecipeStore`, `UserStore`,
  `CollectionStore`) so handlers are testable with `httptest` fakes.
- **Repositories** own all SQL, return domain models, and translate `sql.ErrNoRows`
  into a domain-level "not found".
- **The service layer is optional**: `internal/services` is empty today. Introduce a
  service only when business logic grows beyond a single handler, and keep it consistent
  across the codebase.
- No global mutable state; pass `*sql.DB` and configuration via constructors.

## API Design

### RESTful Conventions
- `GET /resources` — list
- `GET /resources/{id}` — retrieve
- `POST /resources` — create
- `PUT /resources/{id}` — full update
- `DELETE /resources/{id}` — delete

### Request/Response Format
- **Content-Type:** `application/json` for the API (`/api/...`); HTML for web routes.
- Return generic error messages on untrusted paths; never echo raw driver/SQL detail.

### Status Codes
- `200` success · `201` created · `204` no content
- `400` validation · `401` unauthorized · `403` forbidden · `404` not found
- `409` conflict (duplicate) · `500` internal error

### Pagination
- Query params: `?page=1&limit=20`. In SQLite, `OFFSET` requires a preceding `LIMIT`
  (use `LIMIT -1` for "no limit").

## Database Practices
- **Single schema source of truth:** `internal/database/schema.sql`, embedded with
  `go:embed` and applied at startup via `database.ApplySchema(db)`. There is no migration
  tool in the runtime path.
- **Always parameterize:** use `?` placeholders (SQLite) and pass arguments separately.
  Never concatenate user input into SQL.
- **No `RETURNING`, no `AutoMigrate`:** generate UUIDs in Go before INSERT.
- **Transactions:** use `database/sql` transactions (`db.Begin()` → `*sql.Tx`) for
  multi-step writes (e.g. a recipe plus its ingredients, instructions, and tags). Shared
  insert helpers accept a small `sqlExecutor` interface so they run on `*sql.DB` or
  `*sql.Tx`.

## Search & Filtering
- Text search uses SQLite **`LIKE`**.
- Filter server-side with parameterized `WHERE` clauses (difficulty, cook time, ...).
- Build dynamic queries by appending conditions and matching `?` arguments in textual
  order (UPDATEs bind SET args before WHERE args).

## Authentication & Authorization
- JWT signed with HS256; claims carry `user_id` (string) plus standard `exp`/`iat`.
- bcrypt for password hashing; **never** log or serialize the hash (`json:"-"`).
- `appmiddleware` provides the auth middleware and `GetUserID(ctx)`; the owner of a
  resource is taken from the **auth context, never the request body**.
- Foreign resources return `404` (collections, to avoid existence probing) or `403`
  (recipes) per the existing handlers.
- Rate limiting and security headers are applied via middleware.

## Error Handling
- Check every error; wrap with context: `fmt.Errorf("failed to create recipe: %w", err)`.
- Translate `sql.ErrNoRows` into a domain "not found" — never leak it to clients.
- Never return raw driver/SQL error strings on untrusted paths; log server-side and
  return a generic message.

```go
var ErrUserNotFound = errors.New("user not found")

func (r *UserRepository) GetUserByID(id string) (*models.User, error) {
    var u models.User
    err := r.db.QueryRow(`SELECT id, email FROM users WHERE id = ?`, id).
        Scan(&u.ID, &u.Email)
    if errors.Is(err, sql.ErrNoRows) {
        return nil, ErrUserNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("get user by id: %w", err)
    }
    return &u, nil
}
```

## Code Quality
- Prefer clear, explicit code over clever abstractions; use standard-library patterns.
- Raw parameterized SQL is the norm here — there is no query builder to hide behind.
- No global state, no hidden side effects; check every error.
- Keep SQL in repositories, never in handlers.

## Testing
- **Table-driven tests.** Handlers: `net/http/httptest` with interface fakes (see
  `internal/handlers/*_test.go`). Repositories: real SQLite tests (require cgo/gcc).
- Never disable or skip a failing test to make a build pass — fix the root cause.
- Finish every change green: `go build ./...` and `go test ./...` from `backend/`.
  (Repository cgo tests run in CI on Linux even when a local toolchain lacks gcc.)

```go
func TestValidate(t *testing.T) {
    tests := []struct {
        name    string
        input   models.Recipe
        wantErr bool
    }{
        {"valid", models.Recipe{Title: "Soup"}, false},
        {"missing title", models.Recipe{}, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.input.Validate()
            if (err != nil) != tt.wantErr {
                t.Fatalf("Validate() err = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## Security Checklist
- Validate all user input at the boundary (`Validate()` methods).
- Use parameterized queries — no SQL injection.
- Apply rate limiting and CORS for allowed origins.
- Set security headers; serve over HTTPS in production.
- Never log or return secrets, tokens, password hashes, or PII.
