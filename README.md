# OpenCaravan Go

[![CI](https://github.com/opencaravan/opencaravan-go/actions/workflows/ci.yml/badge.svg)](https://github.com/opencaravan/opencaravan-go/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/opencaravan/opencaravan-go.svg)](https://pkg.go.dev/github.com/opencaravan/opencaravan-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/opencaravan/opencaravan-go)](https://goreportcard.com/report/github.com/opencaravan/opencaravan-go)

OpenCaravan Go is the Go protocol package for OpenCaravan, an open protocol for
coordinating group drives over networks.

This module is intentionally small. It defines shared protocol vocabulary,
wire-facing structs, and validation helpers that can be used by servers,
clients, and conformance tests. It does not contain storage engines, server
internals, auth persistence, or deployment tooling.

## Status

OpenCaravan is in early draft. This module starts at `v0` and may change while
Spivot Server and the OpenCaravan specification are built together.

## Install

```bash
go get github.com/opencaravan/opencaravan-go
```

## Usage

```go
package main

import (
	"fmt"
	"time"

	opencaravan "github.com/opencaravan/opencaravan-go"
)

func main() {
	invite := opencaravan.NewJourneyInvitePayload(
		"https://public.spivot.net",
		"journey_123",
		"opaque-token",
		time.Now().Add(30*time.Minute),
	)

	fmt.Println(invite.Type)
}
```

## Package Scope

The package currently includes draft types for:

- server policy and retention capability advertisements
- per-journey policy snapshots
- journey, participant, sharing, and vehicle vocabulary
- portable journey invite payloads
- position telemetry samples

Spivot Server is the reference backend implementation for OpenCaravan and will
consume this module as the protocol surface matures.

## Development

```bash
just ci
```

The module currently has no third-party dependencies.

## License

OpenCaravan Go is licensed under the Apache License, Version 2.0. See
[LICENSE](LICENSE) and [NOTICE](NOTICE) for the full text and attribution.

"OpenCaravan" is a trademark of the OpenCaravan project. The code license does
not grant trademark rights; see [TRADEMARK.md](TRADEMARK.md) for guidance on
naming implementations, services, and forks.
