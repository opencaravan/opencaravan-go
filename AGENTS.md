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

Godoc is a primary product surface for this module, not code commentary.
Two audiences, two bars:

- A developer landing on [pkg.go.dev](https://pkg.go.dev) (or running
  `go doc`) should be able to understand the OpenCaravan vocabulary,
  lifecycle expectations, wire compatibility rules, and privacy
  implications without spelunking through any reference implementation.
- A developer navigating the source should understand *why* each
  unexported helper exists in its current form, especially where the
  obvious naive implementation would be wrong.

Godoc is part of the reader-facing interface. A PR that changes
behavior without updating the affected doc comments is incomplete in
the same way a PR that changes a function without updating its tests
is incomplete.

### Exported symbols (full contract)

Every package and exported const, var, type, field-bearing struct,
function, method, and interface needs a doc comment that:

- Starts with the exported identifier and reads as a complete sentence.
- States the **protocol meaning**: what the symbol represents in the
  OpenCaravan vocabulary, not just its Go shape.
- States **required vs optional** for every field; if absence is
  semantically different from a zero value, say so explicitly.
- States **identifier assignment**: who is allowed to mint a new value
  (server, client, conformance test), when, and via which constructor.
- States the **wire-compatibility contract**: which fields are
  versioned, which are extensible, which are immutable once issued.
- States the **error contract**: which errors are returned when, how to
  detect sentinels via `errors.Is`, what additional context is wrapped.
- States the **concurrency** posture where it differs from "trivially
  safe; immutable after construction."
- Cross-references related symbols via Go 1.19+ `[Identifier]` doc
  links so pkg.go.dev produces clickable references.

Enum-like constants explain the semantic difference between values,
not just the string they serialize to. The wire string is one
implementation detail; the meaning the value commits the issuer to is
the contract.

Package-level documentation in `doc.go` (or per-package equivalent)
uses Go 1.19+ heading syntax (`# Heading`) for a navigable pkg.go.dev
table of contents. For protocol-bearing packages, structure with
sections such as Protocol Model, Identifier Rules, Versioning,
Privacy, and Forward Compatibility.

Add runnable `Example*` functions for the primary usage patterns.
Examples render on pkg.go.dev and are exercised by `go test`, so they
cannot bit-rot silently. Use `// Output:` comments to validate the
deterministic portions. Prefer small, compilable examples that
demonstrate one wire-format-relevant operation at a time.

### Unexported symbols (rationale, not ceremony)

The reader of an unexported function is always reading the source, with
the body in front of them. Document the *why*, not the *what*. The
test for whether a comment is doing real work:

> If a contributor were to delete this unexported symbol in a PR,
> would the surrounding code make clear why that's wrong?

If yes, a short comment or none is fine. If no, the comment is doing
real work and must exist.

Rich docs are warranted on unexported symbols when they encode subtle
invariants the body doesn't show, exist because of a specific bug or
edge case (the obvious naive implementation is wrong), embody a
protocol-level or security policy (validation alphabets, canonical
encodings, hash algorithm choices), or sit at an internal layer
boundary where the correct sequence of primitive operations lives.

Skip ceremonial docs on one-liners where the signature carries the
meaning, pure formatting or conversion helpers, test helpers, and code
that exists purely to reduce duplication with no novel logic.

## Workflow

All local workflows go through [just](https://just.systems/).

```bash
just ci        # format check + tests
just test      # go test ./...
just fmt-check # gofmt check
```

Run `just ci` before pushing. Use conventional commits, sign commits, and keep
PRs focused.
