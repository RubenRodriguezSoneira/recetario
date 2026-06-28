# Go Shared Rule Pack

A condensed, copy-pasteable checklist shared by every agent and skill. It is the short
form of [go.instructions.md](../instructions/go.instructions.md) — when the two differ, the
full instructions win. Load this when you need the rules in context without the prose.

## Must

- ✅ gofmt-clean; imports grouped stdlib / third-party / `recipe-app/...`.
- ✅ Constructor injection (`NewXxx`), no global mutable state.
- ✅ Handlers stay thin: parse → validate → repo/service → respond. No SQL in handlers.
- ✅ Parameterized SQL only (`?` placeholders + args). Never interpolate user input.
- ✅ `defer rows.Close()` on every `Query`; check `rows.Err()` after the loop.
- ✅ Translate `sql.ErrNoRows` into a domain not-found error.
- ✅ Wrap errors with context and `%w`: `fmt.Errorf("failed to X: %w", err)`.
- ✅ Validate models via their `Validate()` method before persisting; map failures to `400`.
- ✅ JWT + bcrypt for auth; secret from `JWT_SECRET` env var.
- ✅ Set `Content-Type` before writing the body; correct status codes.
- ✅ Honor the HTMX branch (`HX-Request: true` → template partial, else JSON).
- ✅ Table-driven tests with `httptest` for handlers.
- ✅ End every change with green `go build ./...` and `go test ./...` (run from `backend/`).

## Must Not

- ❌ ORM / query builder (raw `database/sql` only).
- ❌ String-concatenated SQL containing user input.
- ❌ New dependency without justification + `go mod tidy`.
- ❌ Ignored errors (`_`) without a comment explaining why.
- ❌ Leaking internal/DB errors or secrets/PII to clients or logs.
- ❌ Weakening auth, validation, or the middleware chain.
- ❌ Introducing new architectural patterns when an existing one fits.

## Status Codes (quick reference)

| Code | Meaning |
|------|---------|
| 200 | OK with body |
| 201 | Created |
| 204 | No content |
| 400 | Validation / bad input |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Not found |
| 409 | Conflict (duplicate) |
| 500 | Internal error |

## Layer Boundaries

```
cmd/main.go (wiring) → handlers (HTTP) → services (optional business logic) → repositories (SQL) → database/sql
```

Dependencies point inward. Lower layers never import higher layers. `services/` is currently
empty — only introduce it when logic outgrows a handler, and apply it consistently.
