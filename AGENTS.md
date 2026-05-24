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

## Public API Surface

The exported surface of this module is a protocol commitment and deserves
best-in-class Go API hygiene. The package should feel natural to Go developers
while preserving exact OpenCaravan wire semantics for every implementation.

- Keep exported names small, idiomatic, and durable. Do not export helper types,
  intermediate concepts, or implementation details just because tests or one
  caller can reach them.
- Prefer useful zero values. When a zero value is invalid on the wire, make that
  obvious with validation and Godoc.
- Constructors should earn their place by setting required protocol defaults,
  normalizing input, or preventing malformed values. Otherwise prefer ordinary
  struct literals.
- Use concrete types for protocol concepts. Reach for interfaces only at package
  boundaries where Go callers genuinely need substitution.
- Design marshaling and unmarshaling deliberately. JSON output must be stable,
  predictable, and compatible with non-Go OpenCaravan implementations.
- Implement `json.Marshaler`, `json.Unmarshaler`, `encoding.TextMarshaler`, or
  `encoding.TextUnmarshaler` when a protocol type has canonical text, strict
  validation, version-aware decoding, or compatibility behavior that plain
  struct tags cannot express cleanly.
- Reject malformed wire values with clear errors. Do not silently coerce unknown
  enum values, invalid coordinates, impossible timestamps, or lossy numeric
  representations.
- Preserve unknown or extension fields only through explicit extension points.
  Avoid casual `map[string]any` escape hatches on core protocol types.
- Keep optional fields honest. Prefer `omitempty` only when absence is
  semantically different from a zero value, and use pointer fields when the
  protocol needs that distinction.
- Favor Go-flavored APIs: clear package names, simple methods, sentinel values
  only when useful, wrapped errors at boundaries, table-driven tests, examples,
  and no needless builder/factory ceremony.

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
