# Mineplus Color Logs Fix

## Summary

Mineplus preserves the original ANSI color codes emitted by the Paper server so that INFO/WARN/ERROR lines stay readable inside the proxy output.

## Technical Notes

`TYPE_SER` (server) logs in `lib/errco/errco.go` now forward the raw message returned by `StringGraphic(...)` without forcing gray text. This keeps Paper’s own coloring intact while still funnelling the log through the `[Mineplus]` prefix.

## Updated Binaries

Rebuild the distributions to pick up the change:

- `mineplus-linux-amd64`
- `mineplus-linux-arm64`

(Windows and macOS builds can be produced with the same source change if you distribute them.)

## Deployment

1. Download the latest Mineplus binary for your architecture.
2. Replace the existing `mineplus-proxy` in your server or panel environment.
3. Restart the service – the console will retain the standard Minecraft color palette while staying prefixed with `[Mineplus]`.
