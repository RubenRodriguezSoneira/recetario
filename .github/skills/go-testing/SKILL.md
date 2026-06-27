---
name: go-testing
description: "**WORKFLOW SKILL** — Write or improve Go tests for the Recipe App backend using table-driven tests and net/http/httptest. USE FOR: adding unit tests for models/validation, handler tests with httptest, repository logic tests, and raising coverage. DO NOT USE FOR: generating production code (use go-feature); reviewing the codebase (use the review-codebase agent). INVOKES: file system tools, codebase search, edit, terminal (go test)."
argument-hint: "Describe what to test (a handler, model, repository, or package)"
---

# Go Testing

Write clear, behavior-focused Go tests that follow the existing patterns in the Recipe App
backend and keep the suite green.

## When to Use

- Adding tests for a new or untested handler, model, or repository method
- Reproducing a bug with a failing test before fixing it
- Raising coverage on a package

## Conventions (match the existing suite)

- Test files are `*_test.go` in the same package as the code under test (see
  `internal/handlers/api_test.go`, `internal/models/recipe_test.go`,
  `internal/appmiddleware/auth_test.go`).
- **Table-driven tests** with subtests:
  ```go
  tests := []struct {
      name    string
      input   ...
      wantErr bool
  }{
      {"valid", ..., false},
      {"missing title", ..., true},
  }
  for _, tt := range tests {
      t.Run(tt.name, func(t *testing.T) { /* ... */ })
  }
  ```
- Handlers: use `net/http/httptest`:
  ```go
  req := httptest.NewRequest(http.MethodGet, "/api/recipes", nil)
  rr := httptest.NewRecorder()
  handler.HandleRecipes(rr, req)
  // assert rr.Code, rr.Body, rr.Header()
  ```
  For chi path params, attach a `chi.RouteContext` to the request context and set the param,
  or route through a real `chi.NewRouter()`.
- Test **behavior**, not implementation details. Assert on status code, response body shape,
  and error presence.
- No "Arrange/Act/Assert" boilerplate comments — keep tests self-documenting via names.

## Procedure

### Step 1: Identify the unit
Determine the package and function under test and read it plus its neighbors to understand
inputs, outputs, and error paths.

### Step 2: Load the rules
Skim [go.instructions.md](../../instructions/go.instructions.md) §9 (Testing) and the shared
[Go rule pack](../../references/go-shared-rules.md).

### Step 3: Enumerate cases
List the cases to cover: happy path, each validation failure, each error branch (e.g.
`sql.ErrNoRows`, decode error), and relevant edge cases (empty input, boundary values).

### Step 4: Write the tests
Add a table-driven test per function. Keep each case minimal and independent. Avoid hitting a
real database — exercise handler/validation logic and error mapping. Where a repository needs
a DB, prefer constructor seams already present (note `GetRecipes` returns mock data when
`db == nil`).

### Step 5: Run and iterate
From `backend/`:
```
go test ./...
go test ./internal/<pkg> -run <TestName> -v   # focus a single test while iterating
```
Fix failures (in the test or the code, as appropriate) until green. Never skip or delete a
failing test to go green — fix the root cause.

### Step 6: Summary
- List test files added/modified and the cases covered.
- Confirm `go test ./...` passes.
- Mention any coverage gaps left for follow-up.
