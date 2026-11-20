# Music TUI — Engineering Instructions

## Project

A design-focused streaming music player written in Go.

Stack:
- Go
- Bubble Tea v2
- Lip Gloss v2
- Bubbles v2
- mpv JSON IPC
- SQLite
- TOML configuration

Read `docs/project-spec.md` before implementing major features.

## Cardinal Rules

1. Inspect the repository before making changes.
2. Reuse existing implementations before creating new ones.
3. Every package and file must have a clear responsibility.
4. Follow idiomatic Go conventions.
5. Keep domain, UI, playback, providers, and persistence separate.
6. Never perform blocking operations in Bubble Tea's Update or View functions.
7. Never introduce unnecessary abstractions or dependencies.
8. Prefer standard library solutions where practical.
9. Keep interfaces small and define them where they are consumed.
10. Never ignore meaningful errors.
11. Never hardcode secrets or credentials.
12. Do not use deprecated APIs.
13. Never leave dead or commented-out code.
14. Comments should explain why, not what.
15. Keep changes small, focused, and testable.
16. Never bypass Git hooks or quality checks.
17. Never silently modify unrelated code.
18. Do not claim functionality works without verification.

## Architecture

- `cmd/` contains application entry points.
- `internal/domain/` contains domain models.
- `internal/ui/` contains Bubble Tea components.
- `internal/playback/` contains audio playback logic.
- `internal/providers/` contains streaming integrations.
- `internal/storage/` contains persistence.
- `internal/config/` contains configuration.

Create packages only when needed.

Do not create circular dependencies.

## Go Standards

- Use gofmt.
- Use go vet.
- Write table-driven tests when useful.
- Propagate context for cancellable operations.
- Wrap errors with meaningful context.
- Use structured logging with slog.
- Avoid global mutable state.
- Avoid goroutine leaks.
- Keep functions cohesive.
- Do not use panic for normal error handling.

## Quality Gates

Before completing a task, run:

```sh
gofmt -l .
go vet ./...
go test ./...
```

Also run when applicable:

```sh
go test -race ./...
golangci-lint run
```

If a check fails, investigate and fix the problem.

Never report unexecuted checks as passing.

## UI Standards

- Use Bubble Tea v2.
- Keep View pure.
- Keep network and playback operations asynchronous.
- Support terminal resizing.
- Preserve keyboard-first navigation.
- Use centralized theme values.
- Avoid hardcoded dimensions where practical.
- Handle narrow terminals gracefully.

## Streaming Standards

- Use authorized streaming sources.
- Do not bypass DRM or provider authentication.
- Keep provider-specific code isolated.
- Do not assume all providers support playback.
- Handle unavailable streams gracefully.

## Git Rules

Commit subjects use an area prefix:

[core], [ui], [player], [provider], [storage],
[config], [docs], [test], [build], [release]

Maximum subject length: 72 characters.

No em dashes or attribution trailers.

Never bypass Git hooks.

Never commit or push without explicit authorization.

## Workflow

For each task:

1. Inspect relevant code.
2. Identify the minimal implementation.
3. Explain significant architectural decisions.
4. Implement the change.
5. Add appropriate tests.
6. Run quality checks.
7. Review the diff.
8. Summarize results and limitations.

For large changes, write a short implementation plan first.

Do not stop at scaffolding when implementation was requested.

Do not proceed to the next major milestone without approval.

## Definition of Done

A feature is complete only when:
- It works as intended.
- Errors are handled.
- Tests pass.
- Code is formatted.
- Relevant documentation is updated.
- Architecture boundaries are preserved.
- No unnecessary dependencies were added.

## Git Commit Policy

- Automatically commit completed, tested, and verified features.
- Prefer small, atomic commits after meaningful changes.
- Never commit broken or incomplete implementations.
- Run applicable quality checks before every commit.
- Never bypass Git hooks.
- Follow the project's commit message conventions.
- Never commit secrets, credentials, generated binaries, or temporary files.
- Never amend or rewrite existing commits without permission.
- Never push to GitHub without explicit authorization.
- Do not create commits for trivial intermediate edits.
- If tests fail, fix the issues before committing.
- Report the commit hash and summary after each commit.
