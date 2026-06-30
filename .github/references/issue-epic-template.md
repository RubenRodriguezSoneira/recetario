# Epic Issue Template

Use this template for the **tracking (epic) issue** created by the `plan-to-issues` skill.
Keep it concise: a short summary plus the **Dependencies (Mermaid)** section. Do **not**
embed a manual task-list of sub-issues — they are linked as **native sub-issues** so GitHub
renders progress automatically.

```markdown
## 🎯 Objective
*State the primary goal of this epic and why it matters to the user or product. What problem
are we solving?*

## 📝 Description
*High-level summary of the functionality we are building. Mention the root causes being
fixed and the key technical decisions (1–3 short paragraphs, no filler).*

## 🔗 Related Resources
*   **Sub-issues:** managed as native children of this epic (see the sub-issues panel).
*   **Dependencies:** modelled with native `blocked_by` relationships between sub-issues.
*   **Documentation:** [link to plan / PRD / Notion, if any]

## 🚦 Acceptance Criteria & Success Metrics
- [ ] **Criterion 1:** *e.g., a logged-in user can create a recipe end-to-end.*
- [ ] **Criterion 2:** *...*

## 🛑 Out of Scope
*What we are NOT building in this iteration. What limits the scope of this epic.*

## Dependencies (Mermaid)
```mermaid
graph LR
  A["#<a> ..."] --> B["#<b> ..."]
  B --> C["#<c> ..."]
```
```

## Notes

- Edge `X --> Y` means **Y is blocked by X**. Keep the Mermaid graph in sync with the native
  `blocked_by` links.
- The title should be prefixed with `Epic:` (e.g. `Epic: <Feature> — <short scope>`).
- Reuse existing repo labels (e.g. `enhancement`). Do not invent `epic`/`planning` labels
  unless they already exist.
