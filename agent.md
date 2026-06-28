# OpenCode Global Agent Instructions

## Project Overview

**🍲 Recipe App** is a multi-platform, production-quality learning project focused on building sustainable, maintainable code.

**Core Mission**: Deliver code that prioritizes correctness, clarity, and long-term maintainability above all else.

**Implemented stack & source of truth**: The active component is the Go backend in
`backend/` — **chi v5 + `database/sql` (raw parameterized SQL, no ORM) + SQLite**, with
JWT/bcrypt auth and `Validate()` methods on models. The root `web/` and `mobile/`
directories are placeholders. For code-accurate rules, **`.github/` is authoritative**
(`.github/copilot-instructions.md`, `.github/instructions/go.instructions.md`); see
`backend/agent.md` for the backend summary. `DEVELOPMENT_PLAN.md` is an aspirational
roadmap and does not reflect the current code.

---

## Global Principles

### Code Quality Standards
- **Simplicity first**: Prefer simple, explicit solutions over clever abstractions
- **Readability priority**: Optimize for readability first, performance second
- **Minimal changes**: Make focused, well-scoped changes that address specific requirements
- **Spec adherence**: Follow written specifications exactly—do not assume or guess requirements
- **Explicit over implicit**: Make intentions clear through naming, comments, and structure

### Development Workflow
- **Incremental development**: Build features step-by-step with clear checkpoints
- **Documentation as you go**: Update docs when changing behavior or adding features
- **Version control awareness**: Write commit-ready code with clear change boundaries

---

## Safety & Security Rules

### Critical Constraints (Never Violate)
- ❌ **Never remove or weaken**: Authentication, authorization, or validation logic
- ❌ **Never expose**: Sensitive data (credentials, tokens, PII) in logs or responses
- ❌ **Never perform**: Destructive or schema-altering database operations without review and a backup
- ❌ **Never silently drop**: Fields, API endpoints, or expected behavior without discussion

### Data Handling
- Treat all user data as sensitive by default
- Validate all inputs at system boundaries
- Sanitize data before storage and output
- Follow principle of least privilege for data access

### Database Operations
- Apply schema changes through the single embedded schema
  (`backend/internal/database/schema.sql`, applied via `ApplySchema`)
- Never modify production data without explicit approval
- Back up before destructive operations in development
- Use transactions (`database/sql`) for multi-step data operations

---

## Dependency Management

### Dependency Policy
- **Default stance**: Do not add new dependencies unless explicitly instructed
- **Preference order**:
  1. Built-in language features
  2. Standard library
  3. Existing project dependencies
  4. New well-established libraries (only when justified)

### When Adding Dependencies
- Justify the addition with clear benefits
- Check for maintenance status and security history
- Document the dependency's purpose in package files
- Consider bundle size impact for frontend dependencies

---

## Testing Requirements

### Testing Standards
- **Coverage requirement**: New logic must include corresponding tests
- **Test integrity**: Never disable or skip failing tests to make builds pass
- **Fix don't patch**: Address root causes rather than applying workarounds
- **Test types**: Include unit tests for logic, integration tests for workflows

### Test Quality
- Tests should be readable and self-documenting
- Use descriptive test names that explain what's being tested
- Arrange-Act-Assert pattern for clarity
- Mock external dependencies appropriately

---

## Communication & Collaboration

### Communication Style
- **Be concise**: Provide clear, direct explanations without unnecessary verbosity
- **Explain reasoning**: Clarify the "why" behind non-obvious changes
- **Ask before deciding**: Seek approval before making architectural or structural decisions
- **Provide context**: Include relevant background when proposing changes

### Decision Making
- Flag breaking changes immediately
- Propose alternatives when requirements conflict with best practices
- Document trade-offs when multiple approaches are valid
- Escalate blockers rather than working around them

---

## Code Review Checklist

Before submitting changes, verify:
- [ ] Code follows project style and conventions
- [ ] No security vulnerabilities introduced
- [ ] Tests pass and new tests added where needed
- [ ] Documentation updated if behavior changed
- [ ] No unnecessary dependencies added
- [ ] Changes are minimal and focused
- [ ] Commit messages are clear and descriptive

---

## Error Handling

### Standard Approach
- Handle errors at appropriate abstraction levels
- Provide meaningful error messages for debugging
- Log errors with sufficient context
- Fail fast for developer errors, recover gracefully for user errors
- Never expose stack traces or internal details to end users

---

## Performance Considerations

### Optimization Guidelines
- **Measure before optimizing**: Don't optimize without profiling data
- **Readability first**: Optimize only when performance issues are proven
- **Document trade-offs**: Explain why performance was prioritized over clarity
- **Cache wisely**: Use caching for expensive operations, but keep invalidation simple

---

## Platform-Specific Notes

### Multi-Platform Concerns
- Ensure consistent behavior across the web (HTMX) and Android clients (the Go backend in `backend/` is the active component today)
- Test cross-platform edge cases
- Document platform-specific workarounds clearly
- Maintain feature parity unless explicitly designed otherwise

---

## Questions & Clarifications

When requirements are unclear:
1. **Ask specific questions** rather than making assumptions
2. **Propose options** with trade-offs outlined
3. **Wait for confirmation** before proceeding with architectural changes
4. **Document decisions** made during clarification

---

## Maintenance & Longevity

### Future-Proofing
- Write self-documenting code with clear naming
- Add comments for complex business logic
- Keep configurations external and documented
- Design for easy debugging and troubleshooting

### Technical Debt
- Flag technical debt when it's created
- Propose refactoring opportunities when found
- Balance pragmatism with quality
- Track shortcuts in comments or issues

---
