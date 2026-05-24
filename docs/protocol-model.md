# Protocol Model

OpenCaravan is in early draft. This document explains the current Go package
shape for implementers building servers, clients, or conformance tests against
the protocol vocabulary.

## Usage Example

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
	journeyAvatarImage := &opencaravan.ImageResourceRef{
		ID:           serverAssignedID(),
		Digest:       "sha256:...",
		ContentType:  "image/png",
		WidthPixels:  512,
		HeightPixels: 512,
	}
	journeyBannerImage := &opencaravan.ImageResourceRef{
		ID:           serverAssignedID(),
		Digest:       "sha256:...",
		ContentType:  "image/jpeg",
		WidthPixels:  1200,
		HeightPixels: 400,
	}
	user := opencaravan.User{
		ID: serverAssignedID(),
		Permissions: &opencaravan.UserPermissions{
			InviteGeneration: &opencaravan.InviteGenerationPermissions{
				Scopes:                  []opencaravan.InviteScope{opencaravan.InviteScopeServerRegistration},
				MaxRedemptionsPerInvite: 1,
				MaxLifetimeDays:         30,
			},
		},
		Profile: opencaravan.UserProfile{
			DisplayName: "Riley",
			AvatarImage: userAvatarImage,
			AccentColor: opencaravan.HexColor("#3366cc"),
			Contacts: []opencaravan.UserProfileContact{
				{
					Kind:        opencaravan.UserProfileContactMobileNumber,
					Label:       "Text Riley",
					DisplayText: "+1 503 555 1212",
					Value:       "+15035551212",
				},
				{
					Kind:  opencaravan.UserProfileContactSignal,
					Label: "Signal",
					Value: "https://signal.me/#eu/exampleSignalShareToken",
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
		AvatarImage:     journeyAvatarImage,
		BannerImage:     journeyBannerImage,
		State:           opencaravan.JourneyPlanned,
		DeletionTime:    &deletionTime,
		Features: opencaravan.JourneyFeatures{
			ExportAllowed: true,
			MediaAllowed:  true,
		},
		CreationTime: creationTime,
	}

	coordinator := opencaravan.JourneyParticipant{
		ID:        serverAssignedID(),
		JourneyID: journey.ID,
		UserID:    user.ID,
		Profile:   &user.Profile,
		Privileges: opencaravan.JourneyParticipantPrivileges{
			InviteGeneration: &opencaravan.InviteGenerationPermissions{
				Scopes:                  []opencaravan.InviteScope{opencaravan.InviteScopeJourney},
				MaxRedemptionsPerInvite: 25,
				MaxLifetimeDays:         7,
			},
		},
		JoinTime: creationTime,
	}

	if err := user.Validate(); err != nil {
		panic(err)
	}
	if err := vehicle.Validate(); err != nil {
		panic(err)
	}

	token, err := opencaravan.NewJourneyInviteToken(time.Now().Add(2 * time.Hour))
	if err != nil {
		panic(err)
	}

	invite := opencaravan.NewJourneyInvite(
		"https://public.spivot.net",
		journey.ID,
		token,
		25,
	)
	invite.ID = serverAssignedID()
	invite.CreatedByJourneyParticipantID = coordinator.ID
	invite.CreationTime = creationTime
	invite.PolicyHash = "sha256:..."
	invite.DisplayName = journey.Title
	invite.Links = &opencaravan.JourneyInviteLinks{
		WebURL: "https://public.spivot.net/invites/" + token.Value,
		AppURL: "opencaravan://invite?token=" + token.Value,
	}
	invite.Presentation = &opencaravan.JourneyInvitePresentation{
		Title:       "Join " + journey.Title,
		Summary:     "OpenCaravan journey invite",
		AvatarImage: journey.AvatarImage,
		BannerImage: journey.BannerImage,
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

## Model Notes

Use `NewUUID` when assigning a new protocol object ID in a server,
implementation test, or conformance fixture. Use `ParseUUID` when accepting a
UUID from text, configuration, a command-line flag, or another non-JSON boundary.
The normal client/server wire path is JSON marshaling and unmarshaling.

Servers are invite-governed. `RegistrationInvite` means user registration
requires a server or journey invite with registration scope; `RegistrationClosed`
means the server is not accepting new registrations. Public servers can still be
easy to join by publishing admin-created invites with higher redemption caps
while preserving operator-visible provenance.

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
resources. User profiles, journeys, and vehicles can expose `AvatarImage` for
compact or map representations and `BannerImage` for wider presentation
surfaces. The reference does not carry a URL; clients derive the server fetch
path from the resource ID and use the digest as a cache and integrity key.

`HexColor` is the protocol type for opaque sRGB UI colors such as profile
accent colors. It accepts `#RRGGBB`, rejects alpha, and serializes as canonical
lowercase `#rrggbb`.

`UserProfileContact` stores direct contact identifiers, not actions. Known
contact kinds include `mobile_number`, `email_address`, and `signal`. A
`mobile_number` value can support calling, SMS, or compatible local messaging
apps depending on client capabilities. A `signal` value is an HTTPS Signal.me
link such as `https://signal.me/#p/+15035551212` or
`https://signal.me/#eu/exampleSignalShareToken`. Public web or app links belong
in `UserProfileLink`.

`InviteGenerationPermissions` describes what kind of invites a user or journey
participant may ask the server to generate. Server-scoped user permissions can
grant registration invite powers, while journey participant privileges can grant
journey invite powers with separate redemption-count and lifetime caps.

`User.DeletionAfterInactivityDays` is optional. When set, it declares the number
of inactive days after which a server may delete the user record if no
server-defined activity resets the timer. The day-level unit avoids promising
more scheduling precision than implementations can reliably provide.

`JourneyParticipant` is the membership record for one server-scoped user in one
private journey. A journey participant may carry a profile projection so clients
can render the display name, avatar, accent color, public links, and opt-in
contact methods that are visible to other people sharing the journey.

Journey invites are neutral capability objects. For a one-person private-message
journey invite, set `MaxRedemptions` to `1`. For a link posted to a group chat
or web forum, set `MaxRedemptions` to the server-enforced redemption cap. A
value of `0` means the issuing server has not capped redemptions. `WebURL` is
the browser entry point that lets the server process or redeem the invite;
`AppURL` is the deep link a server or client can use to hand off to a registered
OpenCaravan client app. `JourneyInvitePresentation` can carry title, summary,
avatar, and banner snapshots for rich share surfaces before the client has
fetched the journey.

## Package Scope

The package currently includes draft types for:

- server policy advertisements
- per-journey deletion timestamps and feature flags
- private invite-only journeys, users, journey participants, client apps,
  segments, and vehicles
- in-protocol image resource references for user, journey, and vehicle
  presentation
- invite-governed registration posture and scoped invite generation
  permissions
- portable journey invites with redemption caps, integrity metadata,
  presentation image refs, and web/app link forms
- participant-shared journey media
- position telemetry samples
