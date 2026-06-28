---
name: project-structure
description: "**REFERENCE SKILL** — Authoritative map of the Recipe App Go backend architecture, packages, layering, and conventions. USE FOR: understanding where new code belongs, how layers interact, naming conventions, and the request lifecycle; onboarding to the codebase. DO NOT USE FOR: generating code (use go-feature) or writing tests (use go-testing). Read this before adding or moving files."
argument-hint: "Ask about where code belongs or how a layer works"
---

# Project Structure — Recipe App Go Backend

The canonical description of how the `backend/` Go service is organized. Read this before
creating or moving files so new code lands in the right place and follows existing
conventions.

## Module & Layout

Module path: `recipe-app` (see `backend/go.mod`).

```
backend/
  cmd/
    main.go            # entry point: open DB, create tables, seed, wire chi routes + middleware
  internal/
    appmiddleware/     # chi middleware: auth, CORS, rate limit, error handler, security headers
    handlers/          # HTTP handlers — APIHandler, AuthHandler, UserHandler, WebHandler
    logger/            # structured logger + context helpers (logger.FromContext)
    models/            # domain structs (Recipe, Ingredient, Instruction, User) + Validate()
    repositories/      # data access over database/sql — RecipeRepository, UserRepository
    services/          # business logic — CURRENTLY EMPTY (introduce only when justified)
    storage/           # storage helpers — CURRENTLY EMPTY
    database/          # embedded schema.sql + ApplySchema()
  web/                 # html/template files + static assets served by the Go process
```

## Layering & Dependency Flow

```
cmd/main.go (composition root)
      │ constructs and injects
      ▼
internal/handlers  ──►  internal/services (optional)  ──►  internal/repositories  ──►  database/sql
```

- Dependencies point inward. Lower layers must not import higher layers.
- **Handlers** own HTTP concerns only: parsing, validation calls, status codes, encoding,
  the HTMX branch. No SQL in handlers.
- **Services** hold business logic/orchestration. The package exists but is empty today — do
  not force a service layer; add it only when logic outgrows a single handler, then apply it
  consistently.
- **Repositories** own data access only, returning domain models. Raw `database/sql` with
  parameterized queries — no ORM.
- **Models** are pure domain structs with `json`/`db` tags and explicit `Validate()`.

## Composition Root (`cmd/main.go`)

Everything is wired here: open the SQLite DB (`sql.Open("sqlite", ...)`), `createTables`,
`seedData`, construct repositories and handlers via `NewXxx`, build the chi router, and
register the middleware chain in order:

```
RequestID → Recoverer → RequestLogger → ErrorHandler → CORS → RateLimit → SecurityHeaders
```

Routes are grouped under `/api` (JSON) and `/recipes` (HTML), with auth applied per-route
via `authService.AuthMiddleware` / `OptionalAuthMiddleware`.

## Request Lifecycle (API)

1. chi matches the route and runs the middleware chain.
2. The handler parses params (`chi.URLParam`, `r.URL.Query()`) / decodes JSON.
3. The handler calls `model.Validate()`; failures → `400`.
4. The handler calls a repository method.
5. The repository runs parameterized SQL and returns a domain model or a wrapped error.
6. The handler sets `Content-Type`, encodes JSON (or renders an HTMX template partial when
   `HX-Request: true`), and writes the status code.

## Conventions

- Constructors `NewXxx(deps) *Xxx`; dependencies injected, no global state.
- Errors wrapped with `%w`; `sql.ErrNoRows` translated to not-found.
- Logging via `internal/logger`, request-scoped through `logger.FromContext`.
- Tests are table-driven `*_test.go` beside the code, using `httptest` for handlers.
- Persistence: SQLite (`main.go`); parameterized SQL uses `?` placeholders.

## Where Does New Code Go?

| You are adding… | Put it in… |
|-----------------|------------|
| A new endpoint | `internal/handlers` + route in `cmd/main.go` |
| A DB query/command | `internal/repositories` |
| A domain type or validation | `internal/models` |
| Cross-cutting HTTP behavior | `internal/appmiddleware` |
| Reusable business logic across handlers | `internal/services` (create the file when first needed) |
| Schema change | `internal/database/schema.sql` (applied via `createTables` in `main.go`) |

## Known Debt (respect, don't blindly copy)

- Some handlers take a `ctx interface{}` argument and ignore it in favor of `r.Context()`.
  New code should thread the real `context.Context` instead.
- `CreateRecipe`/`UpdateRecipe` write child rows without a transaction. Prefer wrapping
  related writes in a transaction when you touch them.
