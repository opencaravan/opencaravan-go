# AGENTS.md

OpenCaravan Go is the protocol package for OpenCaravan. It should stay small,
portable, and useful to any Go implementation of the protocol.

## Build & Test

All local workflows go through [just](https://just.systems/).

```bash
just ci        # format check + tests
just test      # go test ./...
just fmt-check # gofmt check
```

Run `just ci` before pushing.

## Code Conventions

- Prefer the Go standard library.
- Do not add third-party dependencies without a clear protocol-level reason.
- Keep this module protocol-focused. Storage engines, server internals, auth
  persistence, and deployment tooling belong in implementations.
- Exported symbols need Godoc comments that start with the symbol name and read
  as complete sentences.
- Use explicit JSON tags on wire-facing structs.
- Avoid committing protocol semantics that are still speculative. Use narrow,
  composable types and validation helpers.

## Git

- Use conventional commits.
- Sign commits.
- Keep PRs focused.
