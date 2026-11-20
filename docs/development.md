# Development

Use Go 1.27.4 or newer and a compatible golangci-lint v2. Run the local quality script from the repository root:

```sh
./scripts/check.sh
```

The script checks formatting, runs vet and unit tests, exercises the race detector, and runs golangci-lint. Go may need network access to download modules before the first build. The tests use local fixtures and require no external services.

Run `go run ./cmd/music-tui` in an interactive terminal for manual testing. Resize the terminal while a track is playing to exercise the sidebar, selected row, and bottom player together.
