# Repository structure

| Path | Responsibility |
| --- | --- |
| `cmd/music-tui` | Program entry point and terminal lifecycle |
| `internal/domain` | Track metadata shared by the catalog, playback, and UI |
| `internal/providers/mock` | Built-in fictional track catalog |
| `internal/playback` | UI-independent simulated playback and queue state |
| `internal/ui` | Bubble Tea state, input, responsive views, and theme |
| `docs` | Project specification, developer documentation, and visual reference |
| `scripts` and `.githooks` | Local quality checks and pre-commit entry point |
| `internal/artwork` | Original terminal-cell cover artwork |
| `internal/config` | Optional XDG TOML configuration and validation |
| `internal/visualizer` | CAVA subprocess, bounded raw frames, and normalization |
| `internal/providers` | Catalog capabilities, registry, safe errors, and mock and YouTube adapters |
