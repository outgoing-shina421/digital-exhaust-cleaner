# Architecture

Digital Exhaust Cleaner is organized as small packages with explicit ownership:

- `cmd/app`: CLI wiring only.
- `internal/config`: YAML configuration loading and validation.
- `internal/logging`: Zap logger construction.
- `internal/scanner`: recursive traversal, filtering, and worker scheduling.
- `internal/metadata`: metadata extraction and streaming hashes.
- `internal/storage`: SQLite schema initialization and persistence.
- `internal/dedupe`: exact duplicate grouping.
- `internal/similarity`: perceptual image hashing and near-duplicate grouping.
- `internal/classifier`: local semantic rules for screenshots, archives, and installers.
- `internal/intelligence`: behavioral heuristics for repeated downloads and abandoned projects.
- `internal/recommendation`: explainable cleanup scoring.
- `internal/cleanup`: reversible quarantine and restore operations.
- `internal/ui`: standalone report rendering and loopback-only interactive cleanup UI.
- `tests`: black-box regression and benchmark coverage using public package APIs.

The engine is local-first. File contents are only read on disk for metadata and hash extraction. No network or telemetry behavior is implemented.

## Safety Model

Recommendations are advisory. Cleanup uses quarantine, restore, and JSON history; hard deletion is intentionally absent from the public API. The interactive UI only binds to loopback addresses and rejects quarantine requests outside the scanned root.

## Future Phases

The current desktop surface is a static local HTML report. Wails can later package the same analysis APIs into a native shell.
