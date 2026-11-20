# Project: Design-Focused Streaming Music TUI in Go

You are a senior Go engineer, terminal UI architect, and experienced open-source maintainer.

Your task is to design and implement a production-quality, keyboard-first, streaming music player for the terminal, written in Go.

The application should feel like a polished desktop music client while remaining a genuine terminal application.

The visual inspiration is the RYOTUNES music player reference image provided with this project.

**This is a real application, not a UI-only demonstration.**

You are responsible for architecture, implementation, testing, documentation, developer experience, and packaging.

Work autonomously, make reasonable technical decisions, and implement the project incrementally. Do not stop after generating scaffolding or a plan.

However, never sacrifice maintainability, correctness, or code quality for implementation speed.

---

# 1. Project Vision

Build a beautiful, responsive, terminal-based streaming music player featuring:

- Modern terminal UI.
- Keyboard-first navigation.
- Online music search.
- Streaming audio playback.
- Persistent playback controls.
- Playlists and favorites.
- Recently played tracks.
- Album artwork where supported.
- Audio visualization.
- Configurable themes.
- Persistent application settings.
- Responsive terminal layouts.

The application should feel fast, elegant, and thoughtfully designed.

Visual inspiration:

- Dark, warm, retro-inspired color palette.
- Cream and muted gold typography.
- Subtle borders and separators.
- Compact sidebar navigation.
- Large featured content areas.
- Carefully aligned text and controls.
- Persistent bottom playback bar.
- Album artwork.
- Minimal visual clutter.

Do not blindly copy the reference image pixel for pixel.

Reinterpret the design using terminal-native UI patterns while preserving its overall visual identity.

The final result should feel intentionally designed for the terminal, not like a web interface awkwardly translated into text.

---

# 2. Technology Stack

Primary language:

- Go, using the current stable release.

Terminal interface:

- Bubble Tea v2.
- Lip Gloss v2.
- Bubbles v2.

Audio playback:

- mpv, controlled through JSON IPC.
- Use a dedicated playback service.
- Never run audio playback directly inside UI rendering functions.

Persistence:

- SQLite.
- Use a maintained SQLite driver.
- Prefer a pure-Go implementation when practical.

Configuration:

- TOML.
- XDG-compliant configuration directories.

Logging:

- Go's standard `log/slog`.

Testing:

- Go's standard testing package.
- Race detector where applicable.
- Integration tests for important boundaries.

Build and automation:

- Go modules.
- Taskfile, if it provides meaningful value.
- GitHub Actions.
- GoReleaser when release packaging is implemented.

Prefer Go's standard library unless an external dependency provides substantial value.

Before choosing dependencies, verify their current APIs, compatibility, maintenance status, and licensing.

Do not use deprecated Bubble Tea APIs or outdated import paths.

---

# 3. Cardinal Engineering Rules

These rules are non-negotiable.

## Rule 1: Organization Is the Point

Every file and directory must have one clear responsibility.

Before creating a file:

1. Inspect the existing repository.
2. Search for existing implementations.
3. Reuse appropriate existing code.
4. Extend existing packages when reasonable.
5. Create new packages only when a meaningful architectural boundary exists.

Never create duplicate implementations of the same responsibility.

Avoid unnecessary abstractions, redundant wrappers, and excessive package fragmentation.

Maintain `docs/structure.md` explaining the repository organization.

The repository structure must reflect the application's actual architecture.

## Rule 2: One Concern Per File

Each file should have a focused responsibility.

Examples:

- `sidebar.go` handles sidebar presentation.
- `player.go` handles player component presentation.
- `theme.go` defines theme styling.
- `mpv.go` handles mpv communication.
- `provider.go` defines streaming provider behavior.

Do not create enormous files containing unrelated logic.

However, do not split trivial functions into separate files merely to satisfy this rule.

Prefer cohesive packages and clear boundaries over arbitrary file-count rules.

## Rule 3: The Repository Is the Source of Truth

All application behavior, defaults, documentation, and build instructions must originate from the repository.

Never depend on undocumented manual changes to a developer's machine.

Application configuration must be loaded from documented locations.

Generated artifacts must be reproducible.

Never commit machine-specific paths, credentials, authentication tokens, or personal configuration.

## Rule 4: Separation of Concerns

Keep these responsibilities independent:

- Terminal rendering.
- User input.
- Application state.
- Music catalog operations.
- Streaming provider integrations.
- Playback engine.
- Persistence.
- Configuration.
- Image rendering.

UI components must not directly perform network requests, database queries, or audio operations.

Use explicit application messages, commands, or service interfaces to communicate across boundaries.

The domain layer must not depend on Bubble Tea.

## Rule 5: Follow Idiomatic Go

Use conventional Go naming and package organization.

Prefer:

- Small interfaces.
- Explicit dependencies.
- Constructor-based dependency injection where needed.
- Clear error handling.
- Context propagation.
- Standard library primitives.
- Composition over inheritance-like abstractions.

Avoid:

- Global mutable application state.
- Unnecessary reflection.
- Giant interfaces.
- Premature generic abstractions.
- Java-style service architecture.
- Circular dependencies.
- Excessive use of `init()`.
- Hidden side effects.

Do not create interfaces for every struct.

Introduce interfaces where they establish a real boundary, enable multiple implementations, or improve testing.

## Rule 6: Comments Explain Why

Code should explain what it does.

Comments should explain non-obvious design decisions, constraints, and tradeoffs.

Never leave commented-out implementations.

Remove dead code.

Document exported APIs when appropriate.

Avoid files dominated by explanatory comments.

## Rule 7: Errors Must Be Handled

Never silently ignore meaningful errors.

Wrap errors with useful context.

Use `errors.Is` and `errors.As` where appropriate.

Do not panic for recoverable failures.

Network failures, unavailable providers, missing audio backends, and malformed configuration must produce useful messages.

The application must not crash simply because a stream fails to load.

## Rule 8: Concurrency Must Be Controlled

Use goroutines only when necessary.

Every long-running goroutine must have a clear owner and shutdown mechanism.

Use contexts for cancellation.

Avoid goroutine leaks.

Never block the Bubble Tea update loop with network requests or playback operations.

Avoid uncontrolled polling.

Prefer event-driven updates where practical.

Use synchronization primitives only when required.

Ensure clean application shutdown.

## Rule 9: Security and Privacy

Never hardcode credentials or authentication tokens.

Never log secrets.

Never store sensitive credentials in plaintext when a suitable secure mechanism exists.

Validate external inputs.

Use parameterized SQL queries.

Avoid shell command construction from untrusted input.

Use `exec.CommandContext` with explicit arguments when launching external processes.

Treat streaming metadata and provider responses as untrusted data.

Respect the user's privacy.

Do not implement telemetry without explicit opt-in.

## Rule 10: Always Pass Quality Checks

Every meaningful change must pass applicable checks:

- `gofmt`
- `go vet`
- `go test ./...`
- `go test -race ./...` where supported
- `golangci-lint run`, once configured

Add project-specific checks when necessary.

Do not bypass Git hooks.

Never use `--no-verify`.

Do not remove or weaken tests to make a change pass.

Fix the underlying issue.

If a check cannot run because of an environmental limitation, report that limitation clearly.

## Rule 11: Git Discipline

Use conventional, focused commits.

Commit subjects must:

- Begin with an area label.
- Be 72 characters or fewer.
- Use imperative wording.
- Have no trailing period.
- Contain no em dash.

Allowed labels:

- `[core]`
- `[ui]`
- `[player]`
- `[provider]`
- `[storage]`
- `[config]`
- `[docs]`
- `[test]`
- `[build]`
- `[release]`

Example:

`[ui] Add responsive sidebar navigation`

For user-visible changes, include a plain-language trailer:

`Note: New: Added keyboard navigation for the music library`

Do not add AI authorship trailers or unrelated attribution.

Never rewrite published history without explicit permission.

Never push to a remote repository unless explicitly authorized.

## Rule 12: Changes Must Reach Users

Features must work in both development and installed builds.

Never introduce behavior that depends on the source checkout being present.

Ensure required assets and default configuration are embedded or packaged correctly.

Support upgrades without destroying user settings or libraries.

Database schema changes must use migrations.

Preserve backward compatibility when practical.

Document breaking changes.

Build and release artifacts must be reproducible.

---

# 4. Streaming Architecture

The application must support streaming providers through a clean abstraction.

Do not tightly couple the UI to a specific music service.

Define domain models for:

- Track.
- Artist.
- Album.
- Playlist.
- Search results.
- Playback state.
- Playback queue.
- Provider capabilities.

Start with a mock provider for development.

Then implement at least one real, authorized streaming integration.

## Implemented catalog milestone

Milestone 2.0 ships YouTube Data API v3 catalog discovery alongside the offline Mock provider. It adds explicit search, batched metadata, paginated results, public playlists, temporary thumbnail display, and browser opening. Native audio playback remains a separate milestone. See [Music sources](streaming.md) for implemented behavior.

## Provider Requirements

The provider abstraction should support capabilities such as:

- Searching tracks.
- Retrieving track metadata.
- Retrieving albums and playlists when supported.
- Obtaining authorized playable media when supported.

Do not assume every provider supports every operation.

Use capability-aware interfaces or explicit unsupported-operation errors.

Keep provider-specific authentication, API responses, and transport details out of the domain layer.

## Streaming Restrictions

Do not assume YouTube Music provides an official third-party audio playback API.

Do not implement DRM circumvention, authentication bypasses, unauthorized stream extraction, or techniques that violate provider access requirements.

If a provider does not support authorized third-party playback, document the limitation.

Prefer a working integration with an authorized streaming source over a fragile unofficial integration.

Do not claim that streaming works unless actual playback has been verified.

---

# 5. Playback Engine

Use mpv as the initial playback backend.

Communicate through JSON IPC.

Implement:

- Play.
- Pause.
- Resume.
- Stop.
- Seek.
- Next track.
- Previous track.
- Volume control.
- Mute.
- Playback position.
- Duration.
- Playback status.
- Error reporting.
- Queue management.

The playback service must manage the mpv process lifecycle.

Handle:

- mpv not installed.
- IPC connection failure.
- Invalid media URL.
- Stream unavailable.
- Network interruption.
- End of track.
- Application shutdown.

Keep the playback engine independent of Bubble Tea.

Design the interface so another playback backend could be introduced later without rewriting the UI.

---

# 6. Terminal UI Architecture

Use Bubble Tea's Model, Update, and View architecture.

Keep rendering deterministic and free of side effects.

The UI should consist of reusable components.

Core components:

- Application shell.
- Sidebar.
- Search bar.
- Home screen.
- Search results.
- Track list.
- Playlist view.
- Album view.
- Now-playing panel.
- Persistent playback bar.
- Status and error messages.
- Help overlay.

Each component should own only the state necessary for its behavior.

Application-wide state should be managed explicitly.

Use Bubble Tea commands for asynchronous operations.

Do not initiate network requests or playback commands from `View()`.

## Keyboard Navigation

Implement intuitive keyboard controls:

| Key | Action |
|---|---|
| `j` / `k` | Navigate lists |
| Arrow keys | Navigate |
| `Enter` | Select or activate |
| `Space` | Play/pause |
| `/` | Focus search |
| `Tab` | Change focus |
| `Esc` | Go back or close overlay |
| `n` | Next track |
| `p` | Previous track |
| `+` / `-` | Volume |
| `?` | Show shortcuts |
| `q` | Quit |

Keyboard shortcuts must respect focused text inputs.

Never trigger global playback shortcuts while the user is typing into a search field.

Provide visible focus indicators.

---

# 7. Visual Design System

The application must follow a consistent design system.

Create a centralized theme definition.

Use semantic colors rather than hardcoding color values throughout UI components.

Suggested palette:

- Background: near-black brown.
- Surface: dark charcoal brown.
- Primary text: warm cream.
- Secondary text: muted beige.
- Accent: soft gold.
- Selected state: muted amber.
- Borders: low-contrast brown.
- Error: muted red.
- Success: soft green.

Use consistent spacing, padding, borders, and typography.

Avoid excessive decorative characters.

Maintain visual hierarchy.

The interface should be beautiful even without terminal graphics support.

## Responsive Layouts

Support:

- Large terminals.
- Medium terminals.
- Compact terminals.

At smaller widths:

- Collapse or simplify the sidebar.
- Reduce nonessential metadata.
- Preserve playback controls.
- Avoid text overflow.
- Avoid broken borders.

Use the terminal's reported dimensions rather than fixed layout assumptions.

## Album Artwork

Support terminal image rendering where practical.

Prioritize Kitty graphics protocol.

Provide a clean fallback for terminals without image support.

Image rendering must be optional and must not break the rest of the application.

Do not introduce album artwork until the core layout and playback functionality are stable.

## Audio Visualization

Implement an optional audio visualization component.

Do not present random animation as real audio analysis.

If actual audio analysis is unavailable, use an explicitly labeled decorative animation or disable the visualization.

Avoid CPU-intensive rendering loops.

Provide configurable animation refresh rates.

---

# 8. Persistence

Use SQLite for local application data.

Potential tables:

- Tracks.
- Playlists.
- Playlist tracks.
- Favorites.
- Playback history.
- Queue snapshots.
- Cached metadata.

Only persist data that the application genuinely needs.

Avoid duplicating provider-owned information unnecessarily.

Use migrations for schema changes.

Store user data in XDG-compliant directories.

Example:

- Config: `~/.config/music-tui/`
- Data: `~/.local/share/music-tui/`
- Cache: `~/.cache/music-tui/`

Respect the corresponding XDG environment variables.

Do not overwrite user configuration during upgrades.

---

# 9. Configuration

Use TOML for user-facing configuration.

Support settings such as:

- Theme.
- Default volume.
- Preferred provider.
- Keybindings.
- Artwork rendering.
- Animation preferences.
- Logging level.
- Playback preferences.

Provide sensible defaults.

Validate configuration at startup.

Report invalid settings clearly.

Support a documented configuration file format.

Avoid introducing configuration options before they are actually used.

---

# 10. Repository Structure

Begin with a minimal structure.

Expand it as implementation requires.

Suggested target structure:

```text
music-tui/
├── cmd/
│   └── music-tui/
│       └── main.go
├── internal/
│   ├── app/
│   ├── domain/
│   ├── ui/
│   │   ├── components/
│   │   ├── screens/
│   │   └── styles/
│   ├── providers/
│   │   └── mock/
│   ├── playback/
│   ├── storage/
│   └── config/
├── docs/
│   ├── architecture.md
│   ├── structure.md
│   ├── development.md
│   └── streaming.md
├── .github/
│   └── workflows/
├── .githooks/
├── go.mod
├── README.md
├── CONTRIBUTING.md
└── LICENSE
```

This is a suggested architecture, not permission to create unnecessary empty directories.

Start with only the packages needed for the current milestone.

Do not introduce a `pkg/` directory without a genuine public-library requirement.

---

# 11. Testing Strategy

Write tests for important behavior.

Prioritize:

- Application state transitions.
- Keyboard navigation.
- Playback queue logic.
- Provider response handling.
- Configuration validation.
- Database operations.
- Error handling.
- Terminal layout calculations.

Use mock providers and fake playback implementations in tests.

Do not require real streaming credentials for ordinary unit tests.

Use integration tests for mpv IPC where practical.

Keep integration tests isolated and deterministic.

Do not make tests depend on live streaming services unless explicitly configured as optional integration tests.

---

# 12. Documentation

Maintain useful documentation throughout development.

README must include:

- Project overview.
- Features.
- Screenshots when available.
- Requirements.
- Installation.
- Development setup.
- Keyboard shortcuts.
- Configuration.
- Supported streaming providers.
- Known limitations.

`docs/architecture.md` must explain major components and data flow.

`docs/structure.md` must explain directory responsibilities.

`docs/development.md` must explain local development and quality checks.

`docs/streaming.md` must explain provider integration, authorization requirements, and playback limitations.

Keep documentation synchronized with implementation.

Never document unimplemented features as working.

---

# 13. Development Milestones

Implement the project incrementally.

## Milestone 1: Terminal UI Foundation

Deliver:

- Working Bubble Tea v2 application.
- Responsive terminal layout.
- Sidebar navigation.
- Home screen.
- Mock music library.
- Track selection.
- Persistent playback bar.
- Keyboard shortcuts.
- Centralized theme.
- Simulated playback state.
- Basic tests.

Acceptance criteria:

- Application launches successfully.
- UI renders without layout corruption.
- Keyboard navigation works.
- Tracks can be selected.
- Playback state changes correctly.
- Terminal resizing works.
- Tests pass.

## Milestone 2.0: YouTube Catalog

Delivered the official YouTube Data API v3 catalog adapter, provider selection, explicit search, metadata enrichment, thumbnails, and public playlists. YouTube videos open in a browser.

## Next milestone: Authorized Native Audio Playback

Deliver:

- mpv integration.
- IPC communication.
- Play/pause/seek.
- Volume control.
- Playback events.
- Queue management.
- Error handling.

Acceptance criteria:

- A supported authorized remote audio URL plays successfully.
- Playback can be controlled from the TUI.
- Playback state stays synchronized with mpv.
- mpv shuts down cleanly.

## Later milestone: Authorized Streaming Playback Provider

Deliver:

- Provider interfaces.
- Real authorized streaming integration.
- Online search.
- Track metadata.
- Search result navigation.
- Loading and error states.

Acceptance criteria:

- User can search an actual provider.
- Search results appear in the TUI.
- An authorized playable track can be selected.
- Playback starts through the playback service.
- Provider failures are handled gracefully.

## Milestone 4: Library and Persistence

Deliver:

- SQLite integration.
- Favorites.
- Playlists.
- Recently played.
- Playback history.
- Persistent preferences.

## Milestone 5: Visual Polish

Deliver:

- Terminal artwork rendering.
- Refined themes.
- Responsive layouts.
- Improved animations.
- Optional audio visualization.
- Better focus states.
- Improved loading indicators.

## Milestone 6: Release

Deliver:

- Automated builds.
- Linux binaries.
- Release archives.
- Installation documentation.
- GitHub Actions.
- GoReleaser configuration.
- Optional Arch Linux packaging instructions.

Do not claim a milestone is complete until its acceptance criteria are satisfied.

---

# 14. Working Rules for Codex

Before making changes:

1. Inspect the repository.
2. Read existing project instructions.
3. Understand current architecture.
4. Identify the smallest necessary change.
5. Check whether functionality already exists.
6. Plan implementation boundaries.

During implementation:

1. Make focused changes.
2. Keep code idiomatic.
3. Avoid unrelated refactoring.
4. Add or update tests.
5. Keep documentation accurate.
6. Run appropriate quality checks.
7. Review the resulting diff.

After implementation:

1. Explain what changed.
2. Explain important architectural decisions.
3. Report tests executed.
4. Report failures or limitations.
5. Identify remaining work.
6. Recommend the next milestone.

Never fabricate test results.

Never silently skip failed checks.

Never introduce placeholder implementations and describe them as complete.

Do not install system packages, modify global machine configuration, delete user data, push commits, or publish releases without explicit authorization.

Do not commit automatically unless requested.

---

# 15. Additional Quality Requirements

The application must:

- Start quickly.
- Avoid unnecessary allocations during rendering.
- Avoid excessive CPU usage.
- Handle terminal resize events.
- Restore terminal state on exit.
- Handle SIGINT and graceful shutdown.
- Avoid memory and goroutine leaks.
- Handle missing dependencies gracefully.
- Work without network access for local UI navigation.
- Provide useful error messages.
- Avoid blocking the UI during slow network operations.

Optimize only after identifying meaningful performance issues.

Favor readability over cleverness.

---

# 16. Definition of Done

A feature is complete only when:

- Its behavior works.
- Its errors are handled.
- Relevant tests exist.
- Formatting and static checks pass.
- Documentation is updated where necessary.
- It does not violate architectural boundaries.
- It does not introduce unnecessary dependencies.
- It works in the intended runtime environment.

A visually convincing screenshot alone does not qualify as a completed feature.

---

# 17. Initial Execution Instructions

Begin by inspecting the repository and available development environment.

Then:

1. Propose a concise architecture appropriate for Milestone 1.
2. Verify the current Bubble Tea v2, Lip Gloss v2, and Bubbles v2 APIs.
3. Initialize the Go module if necessary.
4. Create only the required directories and files.
5. Implement the initial application shell.
6. Add the sidebar and central music area.
7. Add the persistent bottom playback bar.
8. Add mock track data.
9. Implement keyboard navigation.
10. Implement simulated playback state.
11. Add responsive layout handling.
12. Add relevant tests.
13. Run quality checks.
14. Document the implemented architecture.

Use the supplied visual reference to guide the design.

The first milestone should result in a working application that can be launched with:

`go run ./cmd/music-tui`

Do not stop after creating the initial file structure.

Complete Milestone 1, verify its acceptance criteria, and provide a concise implementation report.

After Milestone 1, stop before beginning Milestone 2 so the UI and architecture can be reviewed.

The ultimate objective is a maintainable, beautiful, functional streaming music TUI built with idiomatic Go—not a disposable prototype.
