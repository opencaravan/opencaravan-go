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
- Use explicit JSON tags on wire-facing structs.
- Avoid committing protocol semantics that are still speculative. Use narrow,
  composable types and validation helpers.

## Godoc Expectations

Godoc is a primary product surface for this module. Treat documentation with the
same care as wire compatibility: downstream implementers should be able to read
the generated package docs and understand the OpenCaravan vocabulary without
digging through Spivot Server internals.

- Every package needs a package comment.
- Every exported const, var, type, field-bearing struct, function, method, and
  interface needs a Godoc comment.
- Comments must start with the exported identifier and read as complete
  sentences.
- Comments should explain protocol meaning, lifecycle expectations, privacy or
  interoperability implications, and valid usage. Do not merely restate the type
  signature.
- Enum-like constants should document the semantic difference between values,
  not just expand the identifier.
- Wire-facing structs should document whether fields are required, optional,
  server-assigned, client-assigned, stable identifiers, policy snapshots, or
  extension points.
- Examples are welcome when they make protocol usage clearer. Prefer small,
  compilable examples that would render well on pkg.go.dev.

## Git

- Use conventional commits.
- Sign commits.
- Keep PRs focused.
