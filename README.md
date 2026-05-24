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

Most OpenCaravan object IDs are assigned by the server that owns the object. A
client normally receives those IDs by unmarshaling server responses, then passes
them back when creating related protocol objects. For example, an invite refers
to an existing journey; it does not create the journey.

```go
package main

import (
	"fmt"
	"time"

	opencaravan "github.com/opencaravan/opencaravan-go"
)

func main() {
	journey := opencaravan.Journey{
		ID:    serverAssignedJourneyID(),
		Title: "Sunday Ridge Drive",
	}

	invite := opencaravan.NewJourneyInvitePayload(
		"https://public.spivot.net",
		journey.ID,
		"opaque-token",
		time.Now().Add(30*time.Minute),
	)

	fmt.Println(invite.Type)
}

func serverAssignedJourneyID() opencaravan.UUID {
	id, err := opencaravan.NewUUID()
	if err != nil {
		panic(err)
	}
	return id
}
```

Use `NewUUID` when assigning a new protocol object ID in a server,
implementation test, or conformance fixture. Use `ParseUUID` when accepting a
UUID from text, configuration, a command-line flag, or another non-JSON boundary.
The normal client/server wire path is JSON marshaling and unmarshaling.

## Package Scope

The package currently includes draft types for:

- server policy and retention capability advertisements
- per-journey policy snapshots
- journeys, segments, human participants, client apps, and vehicles
- portable journey invite payloads
- participant-shared journey media
- position telemetry samples

Spivot Server is the reference backend implementation for OpenCaravan and will
consume this module as the protocol surface matures.

## Development

```bash
just ci
```

The module currently has no third-party dependencies.
