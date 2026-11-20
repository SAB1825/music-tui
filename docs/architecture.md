# Architecture

The command passes the mock catalog into the Bubble Tea model. The UI stores screen, focus, selection, filter, and terminal dimensions. The simulation owns the queued track, playback position, play state, and volume. Both consume the same domain track type.

Bubble Tea sends key presses, terminal dimensions, and timer messages to `Update`. A playing simulation schedules a one-second timer through a Bubble Tea command. Changing tracks or restarting playback invalidates any pending tick from the earlier track or play session. Rendering reads the model and applies the shared Lip Gloss theme. Bubbles handles search text editing and the UI filters the in-memory catalog.

The sidebar narrows before collapsing into an inline menu. The library drops metadata columns as space shrinks, and the bottom player folds its controls into fewer lines on short terminals.
