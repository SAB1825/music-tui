# Architecture

The command passes the mock catalog into the Bubble Tea model. The UI stores screen, focus, selection, filter, and terminal dimensions. The simulation owns the queued track, playback position, play state, and volume. Both consume the same domain track type.

Bubble Tea sends key presses, terminal dimensions, and timer messages to `Update`. A playing simulation schedules a one-second timer through a Bubble Tea command. Changing tracks or restarting playback invalidates any pending tick from the earlier track or play session. Rendering reads the model and applies the shared Lip Gloss theme. Bubbles handles search text editing and the UI filters the in-memory catalog.

The sidebar narrows before collapsing into an inline menu. The library drops metadata columns as space shrinks, and the bottom player folds its controls into fewer lines on short terminals.

## Presentation and spectrum

The shell calculates toolbar, sidebar, body, and player dimensions from the window size event. Home chooses a featured layout and discovery grid for the available content rectangle. The navigation cursor and history live in the model, so a narrow window can show the same menu inline. Cover illustrations are rasterized to terminal cells and cached by geometry on resize. Screen changes and exit clear the covers as part of the normal text redraw. Bubble Tea's raw escape command is available, but graphical placements would require separately verified upload, replacement, resize, and cleanup behavior, including under tmux.

The entry point owns CAVA's lifetime. CAVA uses a temporary configuration, a PulseAudio output monitor, and mono ASCII output on stdout. The backend parses complete newline-terminated frames and normalizes levels with peak decay. It keeps at most one pending event and drops a stale frame if the UI falls behind. The UI consumes events with a Bubble Tea command and draws the bars in a pure view. Shutdown cancels and reaps the child and removes its temporary configuration. Backend or configuration failures appear under Settings.


## Mouse and dialogs

The rendering pass records each interactive rectangle while drawing the corresponding control. The model publishes the rendered content and those rectangles as one snapshot. Resize, scrolling, and page changes replace the snapshot together. Bubble Tea v2 delivers click, wheel, and motion events in terminal cells, and pointer motion within an existing target reuses the rendered frame.

A Lip Gloss compositor places one rounded dialog over the application. While it is open, only the dialog publishes targets, and keyboard input goes to the dialog. Settings and Help preserve the underlying focus. Removing a session playlist requires its confirmation action. The playlist data lives in memory.

## Online catalog

The catalog registry resolves an available provider independently of playback. Search, batched video metadata and public playlists are separate capabilities under `internal/providers`. The command registers Mock and YouTube, and injects the chosen entry into the existing model. The YouTube adapter owns endpoint parameters, response mapping, credential handling and bounded retries.

An Enter submission or explicit pagination starts a cancellable Bubble Tea command. Generation checks reject stale replies. The online view keeps its own results and never replaces the mock playback queue. Thumbnails travel through a credential-free loader with network, byte and pixel bounds. The UI requests visible results in small batches, keeps up to 64 cell-rendered covers in the current view and discards them when the view changes.
