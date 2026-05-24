package opencaravan

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Macaroon caveat predicate vocabulary used by OpenCaravan implementations.
//
// The macaroon binary format follows the macaroon spec
// (https://research.google/pubs/macaroons-cookies-with-contextual-caveats-for-decentralized-authorization/);
// OpenCaravan defines only the first-party caveat predicate namespace so any
// implementation that issues, attenuates, or validates session macaroons uses
// the same predicate strings.
//
// Predicate format:
//
//   time<RFC3339         macaroon valid only before the named time
//   journey=UUID         macaroon scoped to a specific journey
//   user=UUID            macaroon scoped to a specific user
//   client_app=UUID      macaroon scoped to a specific client app
//   action=NAME          macaroon permits a specific SessionAction
//
// Predicates are case-sensitive ASCII without surrounding whitespace.

// CaveatKind names a recognized first-party caveat predicate kind.
type CaveatKind string

const (
	// CaveatKindUnknown is the zero value, returned for predicates that do
	// not match any recognized OpenCaravan caveat. Implementations may
	// preserve unknown caveats verbatim and reject the macaroon if any
	// unknown caveat blocks the action being attempted.
	CaveatKindUnknown CaveatKind = ""
	// CaveatKindTimeBefore is a "time<T" caveat.
	CaveatKindTimeBefore CaveatKind = "time_before"
	// CaveatKindJourney is a "journey=UUID" caveat.
	CaveatKindJourney CaveatKind = "journey"
	// CaveatKindUser is a "user=UUID" caveat.
	CaveatKindUser CaveatKind = "user"
	// CaveatKindClientApp is a "client_app=UUID" caveat.
	CaveatKindClientApp CaveatKind = "client_app"
	// CaveatKindAction is an "action=NAME" caveat.
	CaveatKindAction CaveatKind = "action"
)

// Caveat is a parsed first-party caveat predicate.
//
// Only the field corresponding to Kind is populated. Raw contains the
// original predicate string so callers can preserve unknown caveats verbatim
// without round-tripping through structured fields.
type Caveat struct {
	Kind   CaveatKind
	Raw    string
	Time   time.Time
	UUID   UUID
	Action SessionAction
}

// CaveatTimeBefore returns the canonical predicate string asserting the
// macaroon is only valid before t. The time is encoded as UTC RFC3339 with
// nanosecond precision.
func CaveatTimeBefore(t time.Time) string {
	return "time<" + t.UTC().Format(time.RFC3339Nano)
}

// CaveatJourney returns the canonical predicate string scoping a macaroon to
// a specific journey.
func CaveatJourney(id UUID) string {
	return "journey=" + string(id)
}

// CaveatUser returns the canonical predicate string scoping a macaroon to a
// specific user.
func CaveatUser(id UUID) string {
	return "user=" + string(id)
}

// CaveatClientApp returns the canonical predicate string scoping a macaroon
// to a specific client app.
func CaveatClientApp(id UUID) string {
	return "client_app=" + string(id)
}

// CaveatAction returns the canonical predicate string permitting a specific
// SessionAction.
func CaveatAction(action SessionAction) string {
	return "action=" + string(action)
}

// ParseCaveat parses a canonical OpenCaravan caveat predicate string. Unknown
// predicates return a Caveat with Kind == CaveatKindUnknown and Raw set to
// the original string; this is not an error, so callers preserving unknown
// predicates do not need to special-case error handling.
func ParseCaveat(predicate string) (Caveat, error) {
	if predicate == "" {
		return Caveat{}, errors.New("predicate must be non-empty")
	}

	if rest, ok := strings.CutPrefix(predicate, "time<"); ok {
		t, err := time.Parse(time.RFC3339Nano, rest)
		if err != nil {
			return Caveat{}, fmt.Errorf("time caveat value must be RFC3339: %w", err)
		}
		return Caveat{
			Kind: CaveatKindTimeBefore,
			Raw:  predicate,
			Time: t.UTC(),
		}, nil
	}

	key, value, ok := strings.Cut(predicate, "=")
	if !ok {
		return Caveat{Kind: CaveatKindUnknown, Raw: predicate}, nil
	}

	switch key {
	case "journey", "user", "client_app":
		id := UUID(value)
		if !id.Valid() {
			return Caveat{}, fmt.Errorf("%s caveat value must be a valid UUID", key)
		}
		return Caveat{
			Kind: parsedUUIDKind(key),
			Raw:  predicate,
			UUID: id,
		}, nil
	case "action":
		if value == "" {
			return Caveat{}, errors.New("action caveat value must be non-empty")
		}
		// Unknown action values are intentionally accepted so future
		// SessionAction additions round-trip losslessly through implementations
		// running an older protocol version. Callers that need to gate on the
		// action set call SessionAction.Valid() on Caveat.Action.
		return Caveat{
			Kind:   CaveatKindAction,
			Raw:    predicate,
			Action: SessionAction(value),
		}, nil
	default:
		return Caveat{Kind: CaveatKindUnknown, Raw: predicate}, nil
	}
}

func parsedUUIDKind(key string) CaveatKind {
	switch key {
	case "journey":
		return CaveatKindJourney
	case "user":
		return CaveatKindUser
	case "client_app":
		return CaveatKindClientApp
	default:
		return CaveatKindUnknown
	}
}
