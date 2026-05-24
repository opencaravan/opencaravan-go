# AGENTS.md

OpenCaravan Go is the Go protocol package for OpenCaravan. This file is written
for capable agents who can understand the mission, exercise judgment, and leave
the package better than they found it.

## North Star

This module should feel like the obvious Go expression of the OpenCaravan
protocol: small, precise, pleasant to import, and useful to any implementation.
It must serve Go developers without giving Go favored status over the protocol
or over other language implementations.

Spivot Server is the reference implementation, not the owner of the model. When
there is tension between Spivot convenience and protocol clarity, choose the
shape that would still make sense to an independent server, client, or
conformance suite.

## Design Judgment

Start from protocol meaning, then choose the Go shape. The right API should make
OpenCaravan concepts feel inevitable: journeys, participants, vehicles, policy,
invites, and telemetry should be named and documented as protocol vocabulary,
not as artifacts leaked from a database schema or one server's internals.

Prefer boring, durable public surfaces. Export reluctantly; document generously.
Small structs, clear enum-like types, ordinary struct literals, focused
validation methods, and table-driven tests are usually better than elaborate
builder or interface machinery. Constructors should earn their place by setting
required protocol defaults, normalizing input, or preventing malformed values.

Keep this module protocol-focused. Storage engines, server internals, auth
persistence, deployment tooling, and Spivot-only shortcuts belong in
implementations. Do not add third-party dependencies without a clear
protocol-level reason.

## Public API Surface

Every exported name is a promise to downstream implementers. Treat the package
surface as part of the protocol, not just as Go code that happens to serialize
to JSON.

Public APIs should be idiomatic Go: compact names, useful zero values where
possible, concrete types for protocol concepts, interfaces only where callers
genuinely need substitution, clear errors, and no needless ceremony. When a zero
value is invalid on the wire, make that obvious through validation and Godoc.

Wire-facing structs need explicit JSON tags. Optional fields should be honest:
use `omitempty` only when absence is semantically different from a zero value,
and use pointer fields when the protocol needs to preserve that distinction.
Unknown or extension data should flow only through explicit extension points;
avoid casual `map[string]any` escape hatches on core protocol types.

## Marshaling

Marshaling and unmarshaling are protocol behavior. JSON output must be stable,
predictable, and compatible with non-Go OpenCaravan implementations.

Plain struct tags are fine when they fully express the contract. Implement
`json.Marshaler`, `json.Unmarshaler`, `encoding.TextMarshaler`, or
`encoding.TextUnmarshaler` when a type has canonical text, strict validation,
version-aware decoding, or compatibility behavior that default encoding cannot
express cleanly.

Reject malformed wire values with clear errors. Do not silently coerce unknown
enum values, invalid coordinates, impossible timestamps, lossy numeric
representations, or incompatible payload versions. When compatibility requires
tolerance, document the exact tolerance and test it with round-trip and
malformed-input cases.

## Godoc

Godoc is a primary product surface for this module, not commentary on
the source. An implementer of OpenCaravan in another language should
be able to read the generated docs on pkg.go.dev and learn the
protocol vocabulary, the wire-format contracts, identifier-assignment
authority, lifecycle, and privacy posture without opening any Go file.

Exported symbols carry that load. Document the protocol meaning, not
just the Go shape. Package-level docs use Go 1.19+ heading syntax for
navigable structure; cross-reference with `[Identifier]` doc links;
add runnable `Example*` tests for primary usage patterns so the docs
cannot bit-rot.

Unexported symbols answer a different question: *why is this here in
its current form?* The test for whether a comment is earning its place:

> If a contributor deleted this symbol in a PR, would the surrounding
> code make clear why that's wrong?

If no, document the why. If yes, no comment needed.

Godoc is part of the reader-facing interface. A PR that changes
behavior without updating the affected docs is incomplete, the same
way a PR that changes a function without updating its tests is
incomplete.

## Workflow

All local workflows go through [just](https://just.systems/).

```bash
just ci        # format check + tests
just test      # go test ./...
just fmt-check # gofmt check
```

Run `just ci` before pushing. Use conventional commits, sign commits, and keep
PRs focused.
