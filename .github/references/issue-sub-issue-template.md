# Sub-Issue Template

Use this template for each **child (sub-issue)** created by the `plan-to-issues` skill — one
per work item (`todos` row). Keep the body focused: a short problem statement, the approach,
and a verifiable checklist. Reference the epic generically; the real link is the **native
sub-issue relationship**, not prose.

```markdown
Part of <epic> (see tracking issue).

## Bug / Task
*One or two sentences describing the specific problem or unit of work.*

## Fix / Approach
*The intended approach, scoped to this issue only.*

## Done when
- [ ] *Concrete, verifiable outcome 1*
- [ ] *Concrete, verifiable outcome 2*
- [ ] *...*
```

## Notes

- Map to existing labels: `bug` for defects, `enhancement` for features/cleanup/verification.
- Do **not** hand-write "Depends on #N" as the source of truth — dependencies are wired with
  native `blocked_by` relationships (and mirrored in the epic's Mermaid graph). A short prose
  mention is fine for readers, but the API links are authoritative.
- One logical work item per sub-issue; mirror the plan's `todos` granularity.
