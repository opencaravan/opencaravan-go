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
them back when creating related protocol objects. Journeys are private and
invite-only: an invite refers to an existing journey and is created by a journey
participant whose membership can generate invites.

```go
package main

import (
	"fmt"
	"time"

	opencaravan "github.com/opencaravan/opencaravan-go"
)

func main() {
	journey := opencaravan.Journey{
		ID:    serverAssignedID(),
		Title: "Sunday Ridge Drive",
	}

	coordinator := opencaravan.JourneyParticipant{
		ID:            serverAssignedID(),
		JourneyID:     journey.ID,
		ParticipantID: serverAssignedID(),
		Privileges: opencaravan.JourneyParticipantPrivileges{
			CanGenerateInvites: true,
		},
		JoinedAt: time.Now(),
	}

	token, err := opencaravan.NewJourneyInviteToken(
		opencaravan.JourneyInviteMultiUse,
		time.Now().Add(2*time.Hour),
	)
	if err != nil {
		panic(err)
	}
	token.MaxUses = 25

	invite := opencaravan.NewJourneyInvite(
		"https://public.spivot.net",
		journey.ID,
		token,
	)
	invite.ID = serverAssignedID()
	invite.Audience = opencaravan.JourneyInviteGroupAudience
	invite.CreatedByJourneyParticipantID = coordinator.ID
	invite.CreatedAt = time.Now()
	invite.PolicyHash = "sha256:..."
	invite.DisplayName = journey.Title
	invite.Links = &opencaravan.JourneyInviteLinks{
		WebURL: "https://public.spivot.net/invites/" + token.Value,
		AppURL: "opencaravan://invite?token=" + token.Value,
	}
	invite.Integrity = serverSignedInviteIntegrity(invite)

	if err := invite.Validate(); err != nil {
		panic(err)
	}

	fmt.Println(invite.Type)
}

func serverAssignedID() opencaravan.UUID {
	id, err := opencaravan.NewUUID()
	if err != nil {
		panic(err)
	}
	return id
}

func serverSignedInviteIntegrity(invite opencaravan.JourneyInvite) *opencaravan.JourneyInviteIntegrity {
	return &opencaravan.JourneyInviteIntegrity{
		Algorithm: "ed25519",
		KeyID:     "server-key-1",
		Signature: "base64url-signature",
	}
}
```

Use `NewUUID` when assigning a new protocol object ID in a server,
implementation test, or conformance fixture. Use `ParseUUID` when accepting a
UUID from text, configuration, a command-line flag, or another non-JSON boundary.
The normal client/server wire path is JSON marshaling and unmarshaling.

For a one-person private-message invite, use `JourneyInviteSingleUse` and
`JourneyInviteIndividualAudience`. For a link posted to a group chat or web
forum, use `JourneyInviteMultiUse`, optionally capped with `MaxUses`. `WebURL`
is the browser entry point that lets the server process or redeem the invite;
`AppURL` is the deep link a server or client can use to hand off to a registered
OpenCaravan client app.

## Package Scope

The package currently includes draft types for:

- server policy and retention capability advertisements
- per-journey policy snapshots
- private invite-only journeys, segments, human participants, client apps, and
  vehicles
- portable journey invites with single-use and multi-use token semantics,
  integrity metadata, and web/app link forms
- participant-shared journey media
- position telemetry samples

Spivot Server is the reference backend implementation for OpenCaravan and will
consume this module as the protocol surface matures.

## Development

```bash
just ci
```

The module currently has no third-party dependencies.
