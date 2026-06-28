# Copilot Instructions

Global engineering principles for the **Recipe App** Go backend. These apply to every
agent and skill in this repository. Language-specific rules live in
[instructions/go.instructions.md](instructions/go.instructions.md) and the shared
[Go rule pack](references/go-shared-rules.md).

## Engineering Principles

- **Senior-level critical thinking**: Propose the best solution, not the easiest. Identify
  coupling, unnecessary allocations, fragile assumptions, and technical debt before
  presenting a change.
- **Simplicity first**: Prefer simple, explicit code over clever abstractions. Optimize for
  readability first, performance second.
- **Minimal, focused changes**: Make well-scoped changes that address the specific
  requirement. Do not refactor unrelated code in the same change.
- **Spec adherence**: Follow written requirements exactly. When something is ambiguous, ask
  before assuming.
- **Mirror the codebase**: Match the structure, naming, and patterns of existing code. Do
  not introduce new patterns or dependencies unless the existing code has none to follow.

## Project Context

- **What it is**: A multi-platform recipe application. The Go service lives in `backend/`
  and serves both a JSON API (`/api/...`) and server-rendered HTML (chi + `html/template`,
  with HTMX partials).
- **Stack (as actually implemented)**:
  - Go 1.25, router **chi v5**
  - **`database/sql`** with raw, parameterized SQL — **no ORM**
  - Driver: `modernc.org/sqlite` (pure-Go SQLite, used by `main.go`; driver name `"sqlite"`, no CGO)
  - Auth: **JWT** (`golang-jwt/jwt/v5`) + **bcrypt** (`golang.org/x/crypto/bcrypt`)
  - No third-party validation library — models expose `Validate() error` methods
- **Layering (current reality)**: `cmd/main.go` wires chi routes → `handlers` →
  `repositories` → `database/sql`. The `internal/services` and `internal/storage` packages
  exist but are currently empty; introduce a service layer only when business logic grows
  beyond a handler, and keep it consistent across the codebase.

> The repository-level [`agent.md`](../agent.md) and [`backend/agent.md`](../backend/agent.md)
> are higher-level summaries. The rules in `.github/` describe the code **as it exists
> today**. When they conflict, follow `.github/`.

## Safety & Security (never violate)

- Never remove or weaken authentication, authorization, or input validation.
- Never log or return secrets, tokens, password hashes, or PII.
- Never build SQL by concatenating user input — always use `?` placeholders and pass
  arguments separately.
- Never return raw `err.Error()` strings that leak internal/database detail to clients on
  paths that handle untrusted input; log the detail server-side and return a generic message.

## Dependency Policy

Default stance: **do not add dependencies**. Preference order:

1. Go standard library
2. Existing project dependencies (chi, golang-jwt, x/crypto, modernc.org/sqlite, uuid)
3. A new, well-established library — only with explicit justification

If you add one, run `go mod tidy` and explain why in the change description.

## Testing

- New logic must come with tests. Use **table-driven tests** and `net/http/httptest` for
  handlers (the existing pattern in `internal/handlers/api_test.go`).
- Never disable or skip a failing test to make a build pass. Fix the root cause.
- Always finish a change with a green `go build ./...` and `go test ./...` (run from
  `backend/`).

## Error Handling

- Check every error. Never ignore one with `_` unless it is genuinely irrelevant and
  commented as such.
- Wrap errors with context: `fmt.Errorf("failed to create recipe: %w", err)`.
- Translate `sql.ErrNoRows` into a domain-level "not found" rather than leaking it.

## Communication

- Be concise and explain the "why" behind non-obvious changes.
- Flag breaking changes immediately and propose alternatives when a requirement conflicts
  with these rules.
