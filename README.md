# Afterhours Music

Afterhours Music is a keyboard-first terminal music player written in Go. Browse the featured player and discovery cards, search the local catalog, or open the library and session playlists. The bottom player stays visible throughout. The grayscale interface keeps the composition of the [project reference](docs/references/music-ui.png).

## Run locally

Use Go 1.27.4 or newer and an interactive terminal. From the repository root:

```sh
go run ./cmd/music-tui
```

To install the command from this checkout, run `go install ./cmd/music-tui` and launch `music-tui` from your Go binary directory. The sample catalog ships with the binary, and the app runs offline after building.

## Keyboard shortcuts

| Key | Action |
| --- | --- |
| Arrow keys or `h` / `j` / `k` / `l` | Move through cards, tracks, menu items, or filters |
| Enter | Play the selected track, open the selected screen, or confirm the filter |
| `c` | Focus the Home discovery filters |
| `[` / `]` | Go back or forward |
| `u` / `,` | Open Queue or Settings |
| Tab / Shift+Tab | Switch between the menu and track list |
| `/` | Filter local tracks by title, artist, or album |
| Enter in search | Keep the filter and focus the results |
| Esc | Clear search, return to tracks, go back, or close help |
| Space | Play or pause |
| `p` / `n` | Cue the previous or next catalog track |
| Shift+← / Shift+→ | Seek backward or forward five seconds |
| Delete | Ask to remove the open session playlist |
| `+` / `-` | Adjust simulated volume |
| `f` | Favorite or unfavorite the cued track |
| `s` / `r` | Toggle shuffle or repeat for automatic playback |
| `?` | Show keyboard shortcuts |
| `q` / Ctrl+C | Quit. Use Ctrl+C while editing search. |

## Mouse

Click a menu item, filter, track, or card to navigate or select it. Press Enter to play a selected track. In compact layouts, click the **≡ Afterhours** label to open the menu. Scroll the wheel over the sidebar, track area, or an open dialog to scroll that area.

The featured player and bottom dock offer clickable playback controls. Click a point on the progress or volume meter to set its value. Click the search field to type, and click away to keep the current filter. In Kitty, hold Shift while selecting text to use the terminal’s own mouse handling.

Settings and Help open in centered dialogs. Use Tab and Enter to choose an action, or Esc to close. Create a session playlist from the sidebar. It starts with the currently cued track when one is available. Use **Delete list** or the Delete key to remove it after confirmation.

## Optional CAVA spectrum

On Arch Linux, CAVA can be installed with `sudo pacman -S cava`. The app can run without it. Open **Settings** with **,** to see backend or configuration diagnostics.

The default source is the monitor of the default PulseAudio or PipeWire-Pulse output. It can include sound from other applications. Afterhours Music itself still simulates playback silently. A quiet monitor shows an idle spectrum.

To customize the visualizer, create `$XDG_CONFIG_HOME/music-tui/config.toml` or `~/.config/music-tui/config.toml` if `XDG_CONFIG_HOME` is unset. Restart Afterhours Music after editing it.

```toml
[visualizer]
enabled = true
backend = "cava"
bars = 24
framerate = 20
source = "auto"
```

Set `enabled = false` to turn off capture, or use a specific `.monitor` name from `pactl list short sources`. Supported values are 8–64 source bars and 15–25 frames per second. The panel condenses the source bars to fit. The TOML reader is `github.com/pelletier/go-toml/v2`.

## Current scope

The built-in mock catalog is the sole provider. Favorites and recently played tracks last for the session. On Repeat, Night Drive, and Daybreak Tapes are mock collections. Radio and Downloads are marked unavailable.

Album and playlist artwork uses original terminal-cell illustrations. Kitty graphics image placements are deferred because their redraw and cleanup behavior has yet to be verified with Bubble Tea v2 under tmux.

See [development checks](docs/development.md), [architecture](docs/architecture.md), and [repository structure](docs/structure.md).
