---
name: git-commit
description: "**WORKFLOW SKILL** — Create well-formed Conventional Commits for the Recipe App repository. USE FOR: committing staged/unstaged changes, writing commit messages, splitting work into logical commits. DO NOT USE FOR: pushing, opening PRs, or merging. INVOKES: terminal (git), file system tools."
argument-hint: "Optionally describe the change or which files to commit"
---

# Git Commit

Create clear, atomic commits that follow Conventional Commits and the repository's
conventions.

## When to Use

- The user asks to commit changes or "save work".
- After completing a logical unit of work that should be recorded.

## Commit Message Format

```
<type>(<optional scope>): <short imperative description>

<optional body — the "why", wrapped at ~72 chars>

Co-authored-by: Copilot <223556219+Copilot@users.noreply.github.com>
```

- **type**: `feat`, `fix`, `chore`, `refactor`, `test`, `docs`, `build`, `ci`, `perf`.
- **scope** (optional): the area touched, e.g. `backend`, `handlers`, `repositories`,
  `auth`, `web`.
- **description**: imperative mood, lower-case, no trailing period, ≤ ~50 chars.
- **body**: include only when the "why" is not obvious from the description.
- Always append the `Co-authored-by: Copilot` trailer unless the user opts out.

## Procedure

### Step 1: Inspect
```
git status
git --no-pager diff            # unstaged
git --no-pager diff --staged   # staged
```
Understand what changed and group it into logical units. Run a secret scan over the diff
before staging anything sensitive.

### Step 2: Decide commit boundaries
- One logical change per commit. Do not mix unrelated changes.
- If the work spans multiple concerns (e.g. a feature + a dependency bump), make multiple
  commits.

### Step 3: Stage
Stage only the files for the current logical commit:
```
git add <paths>
```
Never stage secrets, local databases (`*.db`), or build artifacts. Check `.gitignore`.

### Step 4: Compose & commit
Write the message per the format above. For multi-line messages, prefer a here-doc-free
approach (pass repeated `-m` flags):
```
git commit -m "feat(handlers): add recipe search endpoint" -m "Co-authored-by: Copilot <223556219+Copilot@users.noreply.github.com>"
```

### Step 5: Verify
```
git --no-pager log -1 --stat
```
Confirm the message, scope, and file list are correct. **Do not push** unless the user asks.

## Constraints

- NEVER commit secrets, credentials, `*.db` files, or build output.
- NEVER mix unrelated changes in one commit.
- NEVER push or open a PR from this skill.
- ALWAYS use imperative mood and a valid Conventional Commit type.
- ALWAYS include the Copilot co-author trailer unless told otherwise.
