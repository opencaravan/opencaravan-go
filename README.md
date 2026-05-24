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
	creationTime := time.Now()
	deletionTime := creationTime.Add(7 * 24 * time.Hour)
	deletionAfterInactivityDays := int64(90)
	userAvatarImage := &opencaravan.ImageResourceRef{
		ID:           serverAssignedID(),
		Digest:       "sha256:...",
		ContentType:  "image/png",
		WidthPixels:  512,
		HeightPixels: 512,
	}
	vehicleAvatarImage := &opencaravan.ImageResourceRef{
		ID:           serverAssignedID(),
		Digest:       "sha256:...",
		ContentType:  "image/png",
		WidthPixels:  512,
		HeightPixels: 512,
	}
	user := opencaravan.User{
		ID: serverAssignedID(),
		Profile: opencaravan.UserProfile{
			DisplayName: "Riley",
			AvatarImage: userAvatarImage,
			AccentColor: "#3366cc",
			Contacts: []opencaravan.UserProfileContact{
				{
					Kind:        opencaravan.UserProfileContactSMS,
					Label:       "Text Riley",
					DisplayText: "+1 503 555 1212",
					URI:         "sms:+15035551212",
				},
			},
		},
		DeletionAfterInactivityDays: &deletionAfterInactivityDays,
	}
	vehicle := opencaravan.Vehicle{
		ID:          serverAssignedID(),
		DisplayName: "Blue Bronco",
		AvatarImage: vehicleAvatarImage,
	}
	journey := opencaravan.Journey{
		ID:              serverAssignedID(),
		OriginServerURL: "https://public.spivot.net",
		Title:           "Sunday Ridge Drive",
		State:           opencaravan.JourneyPlanned,
		DeletionTime:    &deletionTime,
		Features: opencaravan.JourneyFeatures{
			ExportAllowed: true,
			MediaAllowed:  true,
		},
		CreationTime:    creationTime,
	}

	coordinator := opencaravan.JourneyParticipant{
		ID:            serverAssignedID(),
		JourneyID:     journey.ID,
		UserID:        user.ID,
		Profile:       &user.Profile,
		Privileges: opencaravan.JourneyParticipantPrivileges{
			CanGenerateInvites: true,
		},
		JoinTime: creationTime,
	}

	if err := user.Validate(); err != nil {
		panic(err)
	}
	if err := vehicle.Validate(); err != nil {
		panic(err)
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
	invite.CreationTime = creationTime
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

`Journey.DeletionTime` is the immutable scheduled hard-deletion time. A nil
value means the server has not scheduled the journey for deletion. Journey-level
feature flags such as `ExportAllowed` and `MediaAllowed` describe capabilities
that clients can render directly.

`User.ID` is scoped to one server registration. It is not a global person ID and
should not be used to correlate a person across servers. User-controlled client
apps supply and maintain profile information for each server registration; a
server republishes its current view of that profile to authorized journey
participants. Clients may mirror one profile across servers or tailor profile
details for each server.

`ImageResourceRef` is the reusable in-protocol handle for server-accepted image
resources. User profiles and vehicles can both expose `AvatarImage` for compact
or map representations and `BannerImage` for wider presentation surfaces. The
reference does not carry a URL; clients derive the server fetch path from the
resource ID and use the digest as a cache and integrity key.

`User.DeletionAfterInactivityDays` is optional. When set, it declares the number
of inactive days after which a server may delete the user record if no
server-defined activity resets the timer. The day-level unit avoids promising
more scheduling precision than implementations can reliably provide.

`JourneyParticipant` is the membership record for one server-scoped user in one
private journey. A journey participant may carry a profile projection so clients
can render the display name, avatar, accent color, public links, and opt-in
contact methods that are visible to other people sharing the journey.

For a one-person private-message invite, use `JourneyInviteSingleUse` and
`JourneyInviteIndividualAudience`. For a link posted to a group chat or web
forum, use `JourneyInviteMultiUse`, optionally capped with `MaxUses`. `WebURL`
is the browser entry point that lets the server process or redeem the invite;
`AppURL` is the deep link a server or client can use to hand off to a registered
OpenCaravan client app.

## Package Scope

The package currently includes draft types for:

- server policy advertisements
- per-journey deletion timestamps and feature flags
- private invite-only journeys, users, journey participants, client apps,
  segments, and vehicles
- in-protocol image resource references for user and vehicle presentation
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
