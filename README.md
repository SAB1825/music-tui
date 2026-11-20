# Afterhours Music

Afterhours Music is a keyboard-first terminal music player written in Go. Milestone 1 includes a Home screen, a browsable mock library, local search, and an always-visible player with simulated playback. The warm palette and layout take inspiration from the [project reference](docs/references/music-ui.png).

## Run locally

Use Go 1.27.4 or newer and an interactive terminal. From the repository root:

```sh
go run ./cmd/music-tui
```

To install the command from this checkout, run `go install ./cmd/music-tui` and launch `music-tui` from your Go binary directory. The sample catalog ships with the binary, and the app runs offline after building.

## Keyboard shortcuts

| Key | Action |
| --- | --- |
| ↑ / ↓ or `k` / `j` | Move through tracks or menu items |
| Enter | Play the selected track or open the selected screen |
| Tab / Shift+Tab | Switch between the menu and track list |
| `/` | Filter local tracks by title, artist, or album |
| Enter in search | Keep the filter and focus the results |
| Esc | Clear search, return to tracks, or close help |
| Space | Play or pause |
| `p` / `n` | Cue the previous or next catalog track |
| `+` / `-` | Adjust simulated volume |
| `?` | Show keyboard shortcuts |
| `q` / Ctrl+C | Quit. Use Ctrl+C while editing search. |

The player progresses through the catalog while playing and stops at the end. Playback is silent and state lasts for the current session. The built-in catalog is the only source in this milestone. The app requires no configuration.

See [development checks](docs/development.md), [architecture](docs/architecture.md), and [repository structure](docs/structure.md).
