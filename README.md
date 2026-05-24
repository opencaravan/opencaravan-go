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

OpenCaravan is in early draft. This module starts at `v0` and may change while
Spivot Server and the OpenCaravan specification are built together.

## Install

```bash
go get github.com/opencaravan/opencaravan-go
```

## Docs

- [Protocol model](docs/protocol-model.md) describes the current draft objects,
  ID lifecycle, image resources, invites, and a complete usage example.
- [Go Reference](https://pkg.go.dev/github.com/opencaravan/opencaravan-go)
  exposes the package API surface and Godoc.

## Package Scope

The package currently includes draft types for server policy advertisements,
invite-governed registration, private journeys, users, profile projections,
vehicles, segments, participant-shared media, journey invites, and position
telemetry samples.

Spivot Server is the reference backend implementation for OpenCaravan and will
consume this module as the protocol surface matures.

## Development

All local workflows go through [just](https://just.systems/).

```bash
just ci
```

The module currently has no third-party dependencies.
