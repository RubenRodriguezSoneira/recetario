---
name: go-feature
description: "**WORKFLOW SKILL** — Generate new Go code for the Recipe App backend following project rules and the existing chi + database/sql patterns. USE FOR: creating new handlers, repository methods, models, validations, and route wiring; generating code that complies with go.instructions.md. DO NOT USE FOR: writing tests (use go-testing); reviewing existing code (use the review-codebase agent); updating dependencies (use update-deps); generating documentation. INVOKES: file system tools, codebase search, edit, terminal (go build/test)."
argument-hint: "Describe the new feature, endpoint, or component to generate"
---

# Go Feature Development

Generate new Go code for the Recipe App backend that complies with project rules, mirrors
existing patterns, and passes build and tests.

## When to Use

- Adding a new HTTP endpoint or operation
- Adding a repository method backed by `database/sql`
- Adding or extending a model and its `Validate()` method
- Wiring new routes/middleware in `cmd/main.go`

## Procedure

### Step 1: Understand the Request
Clarify: HTTP method + path, request/response shape, auth requirement, which aggregate it
touches, and persistence needs. If vague, ask before proceeding.

### Step 2: Load the Rules
Re-read [go.instructions.md](../../instructions/go.instructions.md) and the shared
[Go rule pack](../../references/go-shared-rules.md) before generating code. They may have
changed since the last run.

### Step 3: Explore Existing Patterns
Study the closest existing code and follow it exactly:
- Handlers: `internal/handlers/api.go`, `auth.go`, `user.go`
- Repositories: `internal/repositories/recipe_repository.go`, `user_repository.go`
- Models: `internal/models/recipe.go`
- Routing/middleware: `cmd/main.go`

Identify the correct package, struct, constructor, and namespace. Do not invent new patterns
or abstractions if the codebase already has one.

### Step 4: Baseline
From `backend/`, run `go build ./...` and `go test ./...`; note the current state so you do
not introduce regressions.

### Step 5: Generate the Code (bottom-up)
1. **Model** — struct with `json`/`db` tags + `Validate() error` (required fields, enums).
2. **Repository** — method on the aggregate's repo struct: parameterized SQL,
   `QueryRow`/`Query` + `defer rows.Close()` + `rows.Err()`, `sql.ErrNoRows` → not-found,
   `fmt.Errorf("...: %w", err)`. Use a transaction for multi-step writes.
3. **Handler** — thin: parse → `Validate()` → repo/service → set `Content-Type` → encode →
   correct status code. Honor the `HX-Request` HTMX branch where the endpoint renders HTML.
4. **Routes** — register under the correct `chi` group in `cmd/main.go` with the right auth
   middleware.

Principles:
- Keep it minimal — only what was requested; no speculative features.
- No new dependencies (if truly required, justify and run `go mod tidy`).
- Make it testable — constructor injection, well-typed returns.
- Never concatenate user input into SQL; never leak internal errors to clients.

### Step 6: Build & Test Loop
From `backend/`, run `go vet ./...`, `go build ./...`, `go test ./...`. Fix failures and
repeat until all green. Do not stop until the loop exits clean.

### Step 7: Summary
- List files created/modified.
- Describe the feature.
- Confirm build passing and tests green.
- Suggest the `go-testing` skill if more coverage is warranted.
