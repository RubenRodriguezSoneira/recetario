# Agents & Skills — Recipe App (Go backend)

AI agent and skill definitions for this repository, adapted to the Go stack
(**chi v5 + `database/sql`, JWT/bcrypt, no ORM**). This mirrors the structure used in our
.NET services but targets the Go backend in `backend/`.

## Layout

```
.github/
├── AGENTS.md                       # this index
├── copilot-instructions.md         # global engineering principles
├── instructions/
│   └── go.instructions.md          # canonical Go coding rules (applyTo: backend/**/*.go)
├── references/
│   └── go-shared-rules.md          # condensed shared checklist
├── agents/
│   ├── implement-feature.agent.md  # build a new operation, or review an existing one
│   ├── review-codebase.agent.md    # full-repo consistency review (report-only until approved)
│   └── update-deps.agent.md        # update Go modules with build/test verification
└── skills/
    ├── go-feature/                 # generate new Go code following project patterns
    ├── go-testing/                 # write table-driven + httptest tests
    ├── project-structure/          # architecture & "where does code go" reference
    └── git-commit/                 # Conventional Commits workflow
```

## Agents

| Agent | Description |
|-------|-------------|
| `implement-feature` | Implement a new endpoint/operation (handler + repository + model + routes) — Mode A: new, Mode B: review existing. |
| `review-codebase` | Validate the whole backend against the rules and conventions; produces a severity-categorized report and correction plan. Does not fix until approved. |
| `update-deps` | Find and update outdated Go modules in `backend/go.mod`, verifying build and tests, then prepare a branch and commit. |

## Skills

| Skill | Description |
|-------|-------------|
| `go-feature` | Generate new Go code (handlers, repositories, models, routes) that complies with the project rules. |
| `go-testing` | Write or improve table-driven tests using `net/http/httptest`. |
| `project-structure` | Authoritative map of packages, layering, request lifecycle, and where new code belongs. |
| `git-commit` | Create well-formed Conventional Commits. |

## How It Fits Together

1. **Global rules** (`copilot-instructions.md`) apply everywhere.
2. **Language rules** (`instructions/go.instructions.md`) auto-apply to `backend/**/*.go`;
   the **shared pack** (`references/go-shared-rules.md`) is the condensed checklist both
   agents and skills load.
3. **Agents** orchestrate multi-step workflows and load the relevant skills/rules.
4. **Skills** are focused, reusable procedures invoked by agents or directly by the user.

## Conventions

- The rules describe the codebase **as it exists today**. The aspirational `agent.md` /
  `backend/agent.md` files mention GORM/PostgreSQL; where they conflict with `.github/`,
  follow `.github/`.
- Run Go commands from `backend/`. Every change ends with green `go build ./...` and
  `go test ./...`.
