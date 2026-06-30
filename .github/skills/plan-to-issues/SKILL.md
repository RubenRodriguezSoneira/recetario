---
name: plan-to-issues
description: "**WORKFLOW SKILL** — Turn a multi-phase plan into a tracked GitHub issue hierarchy for the Recipe App repo. USE FOR: splitting a plan/epic into one tracking issue + sub-issues, attaching native GitHub sub-issue relationships, adding a Mermaid dependency graph, assigning the epic, and wiring native blocked-by dependencies via the GitHub API. Uses epic + sub-issue body templates in references/. DO NOT USE FOR: writing code (use go-feature), committing (use git-commit), or closing/merging issues. INVOKES: terminal (gh CLI / gh api), file system tools."
argument-hint: "Point at the plan (e.g. session plan.md) or describe the work to split into issues"
---

# Plan → GitHub Issues

Turn an approved multi-phase plan into a clean, tracked GitHub issue hierarchy: one
**epic/tracking issue** plus one **sub-issue per work item**, connected by **native GitHub
sub-issue relationships**, a Mermaid dependency graph, and **native "blocked by"
relationships**.

Use the reference templates for issue bodies:

- Epic body → [issue-epic-template.md](../../references/issue-epic-template.md)
- Sub-issue body → [issue-sub-issue-template.md](../../references/issue-sub-issue-template.md)

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
defects, `enhancement` for features/cleanup/verify). Use the
[sub-issue template](../../references/issue-sub-issue-template.md): a short **Bug/Task**, a
**Fix/Approach**, and a **Done when** checklist. Reference the tracking issue generically
("Part of <epic> — see tracking issue"); the authoritative link is the native sub-issue
relationship wired in Step 4, not prose.

```bash
gh issue create -R <owner>/<repo> --label bug \
  --title "Fix: <concise problem>" \
  --body $'Part of <epic> (see tracking issue).\n\n## Bug / Task\n...\n\n## Fix / Approach\n...\n\n## Done when\n- [ ] ...'
```

> ⚠️ **Shell safety**: create issues **one `gh issue create` per command**. Do NOT capture
> output with nested `$(...)` inside another `$(...)` — the shell guard blocks nested
> command substitution. Read the printed URL/number from each call instead.

### Step 3: Create the tracking (epic) issue
Use the [epic template](../../references/issue-epic-template.md). The body holds the
**Objective**, **Description**, **Related Resources**, **Acceptance Criteria**, **Out of
Scope**, and the **Dependencies (Mermaid)** graph. Do **not** hand-write a task-list of
sub-issues — they are linked as native children in Step 4, so GitHub renders progress
automatically. Prefix the title with `Epic:`. For long bodies, write the markdown to a temp
file and pass `--body-file` to avoid shell-escaping issues with fenced code blocks.

```bash
gh issue create -R <owner>/<repo> --label enhancement \
  --title "Epic: <Feature> — <short scope>" \
  --body-file /tmp/epic-body.md
```

Mermaid renders natively in GitHub issue bodies — use a ` ```mermaid ` fenced block,
`graph LR`, node labels quoting the issue number, and `-->` edges for "blocks".

### Step 4: Link sub-issues as native children of the epic
GitHub's sub-issues API attaches each child to the epic so the epic shows a real sub-issue
panel with progress. The POST body needs the child's **internal numeric `id`** (not its
number), passed as a **typed** field named `sub_issue_id` with `-F`.

```bash
# Get the internal id of each child issue:
gh api repos/<owner>/<repo>/issues/<child#> --jq .id

# Attach the child to the epic (one call per child):
gh api --method POST repos/<owner>/<repo>/issues/<epic#>/sub_issues \
  -F sub_issue_id=<child_id>
```

Gotchas:
- The field is `sub_issue_id` (not `issue_id`) for this endpoint → a wrong key returns
  `422 ... missing required key: sub_issue_id`.
- Use `-F` (typed), not `-f` → `-f` sends a string and the API rejects it.
- `sub_issue_id` is the value from `--jq .id`, **not** the issue number.

### Step 5: Assign the epic
Assign the tracking issue to its creator (or the named owner) so ownership is clear:

```bash
gh issue edit <epic#> -R <owner>/<repo> --add-assignee @me
```

### Step 6: Wire native "blocked by" dependencies
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
- This endpoint's field is `issue_id` (the blocker's id) — distinct from `sub_issue_id` in
  Step 4.
- Use `-F` (typed), not `-f` → `-f` sends a string and the API returns
  `422 ... is not of type integer`.
- `issue_id` is the value from `--jq .id`, **not** the issue number.
- Add one blocker per call; an issue can be blocked by several.

### Step 7: Verify
```bash
# Confirm the epic's native sub-issues:
gh api repos/<owner>/<repo>/issues/<epic#>/sub_issues --jq '[.[].number]|sort|join(", ")'

# Per child issue, confirm its blockers:
gh api repos/<owner>/<repo>/issues/<n>/dependencies/blocked_by --jq '[.[].number]|join(", ")'
```
Confirm every sub-issue is attached to the epic, the Mermaid graph renders, the epic is
assigned, and every blocked-by edge matches the plan's `todo_deps`.

## Constraints

- NEVER create duplicate issues — check for an existing epic first.
- NEVER invent labels; reuse the repo's existing label set.
- ALWAYS create sub-issues before the epic so the epic can link real numbers.
- ALWAYS attach sub-issues as native children of the epic (Step 4) — do not rely on a manual
  task-list in the epic body.
- ALWAYS use the reference templates for issue bodies (epic + sub-issue).
- ALWAYS keep the Mermaid graph and the native blocked-by links in sync with each other.
- Keep `blocked_by` direction correct: edge `X --> Y` in the graph means **Y is blocked by X**.
- Do NOT close, reopen, or comment-spam issues from this skill.
