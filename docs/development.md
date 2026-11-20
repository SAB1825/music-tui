# Development

Use Go 1.27.4 or newer and a compatible golangci-lint v2. Run the local quality script from the repository root:

```sh
./scripts/check.sh
```

The script checks formatting, runs vet and unit tests, exercises the race detector, and runs golangci-lint. Go may need network access to download modules before the first build. The tests use local fixtures and require no external services.

Run `go run ./cmd/music-tui` in an interactive terminal for manual testing. Resize the terminal while a track is playing to exercise the sidebar, selected row, and bottom player together.

For presentation review, check 160×45, 120×35, and 80×24. Exercise toolbar search, the collapsed menu, category filters, the discovery cards, and the persistent player. Unit tests use a fake CAVA subprocess and fragmented raw frames. A real capture test requires CAVA and a PulseAudio-compatible output monitor. Use a dedicated test output monitor to keep the fixture separate from everyday audio.
