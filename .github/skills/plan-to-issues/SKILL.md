---
name: plan-to-issues
description: "**WORKFLOW SKILL** — Turn a multi-phase plan into a tracked GitHub issue hierarchy for the Recipe App repo. USE FOR: splitting a plan/epic into one tracking issue + sub-issues, adding task-list links, a Mermaid dependency graph, assigning the epic, and wiring native blocked-by dependencies via the GitHub API. DO NOT USE FOR: writing code (use go-feature), committing (use git-commit), or closing/merging issues. INVOKES: terminal (gh CLI / gh api), file system tools."
argument-hint: "Point at the plan (e.g. session plan.md) or describe the work to split into issues"
---

# Plan → GitHub Issues

Turn an approved multi-phase plan into a clean, tracked GitHub issue hierarchy: one
**epic/tracking issue** plus one **sub-issue per work item**, linked by a task list, a
Mermaid dependency graph, and **native "blocked by" relationships**.

## When to Use

- A plan (e.g. `plan.md` in the session folder, or the SQL `todos` table) is approved and
  the user wants it tracked as GitHub issues for better management.
- The user asks to "split into issues", "make an epic", or "track this as issues".

## When NOT to Use

- Writing or changing code → `go-feature`.
- Committing → `git-commit`.
- Closing, merging, or re-prioritising existing issues.

## Prerequisites

```bash
gh auth status                                   # authenticated
gh repo view <owner>/<repo> --json viewerPermission,hasIssuesEnabled
```

- `viewerPermission` must be `WRITE`/`MAINTAIN`/`ADMIN` and `hasIssuesEnabled: true`.
- Default repo is `TorratDev/recetario`. Confirm the target if ambiguous.
- Check existing labels first; reuse them, don't invent new ones:
  ```bash
  gh label list -R <owner>/<repo>
  ```

## Procedure

### Step 1: Derive the work items
Read the plan source (session `plan.md` and/or the `todos` + `todo_deps` SQL tables). Each
todo becomes one sub-issue. Capture the **dependency edges** (`todo_deps`) — they drive both
the graph and the blocked-by links. Confirm granularity with the user if unsure (one issue
per todo vs. grouped).

### Step 2: Create the sub-issues first
Create each child issue and **record its number**. Map to existing labels (`bug` for
defects, `enhancement` for features/cleanup/verify). Keep bodies focused: a short
**Bug/Task**, a **Fix/Approach**, and a **Done when** checklist. Reference the tracking
issue generically ("Part of the web UI redesign — see tracking issue").

```bash
gh issue create -R <owner>/<repo> --label bug \
  --title "Fix: <concise problem>" \
  --body $'Part of <epic> (see tracking issue). Depends on #<n>.\n\n## Bug\n...\n\n## Fix\n...\n\n## Done when\n- [ ] ...'
```

> ⚠️ **Shell safety**: create issues **one `gh issue create` per command**. Do NOT capture
> output with nested `$(...)` inside another `$(...)` — the shell guard blocks nested
> command substitution. Read the printed URL/number from each call instead.

### Step 3: Create the tracking (epic) issue
Body contains: short context, a **task list linking every sub-issue** (renders progress),
the **Mermaid graph**, and scope notes.

```bash
gh issue create -R <owner>/<repo> --label enhancement \
  --title "<Epic> — tracking" \
  --body $'<context>\n\n## Sub-issues\n- [ ] #<a> — ...\n- [ ] #<b> — ...\n\n## Dependencies\n```mermaid\ngraph LR\n  A["#<a> apply"] --> B["#<b> ..."]\n```\n\n## Scope notes\n- ...'
```

Mermaid renders natively in GitHub issue bodies — use a ` ```mermaid ` fenced block,
`graph LR`, node labels quoting the issue number, and `-->` edges for "blocks".

### Step 4: Assign the epic
Assign the tracking issue to its creator (or the named owner) so ownership is clear:

```bash
gh issue edit <epic#> -R <owner>/<repo> --add-assignee @me
```

### Step 5: Wire native "blocked by" dependencies
GitHub's issue-dependencies API records real blocked-by/blocks links (shown in the sidebar).
The POST body needs the blocker's **internal numeric `id`** (not its number), passed as a
**typed** field with `-F`.

```bash
# Get the internal id of each issue:
gh api repos/<owner>/<repo>/issues/<n> --jq .id

# Add a blocker (issue <n> is blocked by the issue whose id is <blocker_id>):
gh api --method POST repos/<owner>/<repo>/issues/<n>/dependencies/blocked_by \
  -F issue_id=<blocker_id>
```

Gotchas:
- Use `-F` (typed), not `-f` → `-f` sends a string and the API returns
  `422 ... is not of type integer`.
- `issue_id` is the value from `--jq .id`, **not** the issue number.
- Add one blocker per call; an issue can be blocked by several.

### Step 6: Verify
```bash
# Per child issue, confirm its blockers:
gh api repos/<owner>/<repo>/issues/<n>/dependencies/blocked_by --jq '[.[].number]|join(", ")'
```
Confirm the epic's task list and Mermaid graph render, the epic is assigned, and every
blocked-by edge matches the plan's `todo_deps`.

## Constraints

- NEVER create duplicate issues — check for an existing epic first.
- NEVER invent labels; reuse the repo's existing label set.
- ALWAYS create sub-issues before the epic so the epic can link real numbers.
- ALWAYS keep the Mermaid graph and the native blocked-by links in sync with each other.
- Keep `blocked_by` direction correct: edge `X --> Y` in the graph means **Y is blocked by X**.
- Do NOT close, reopen, or comment-spam issues from this skill.
