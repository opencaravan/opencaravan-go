package opencaravan

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"
)

// SessionRequestType is the canonical type value for session request payloads.
const SessionRequestType = "opencaravan.session_request"

// SessionResponseType is the canonical type value for session response payloads.
const SessionResponseType = "opencaravan.session_response"

// SessionVersion is the current session wire protocol version.
const SessionVersion = 1

// SessionAction names an action a session-bound macaroon may permit.
//
// The session granted to a client is the intersection of: (1) the actions the
// client requests, (2) the actions the user/client_app is permitted to
// perform under server and journey policy, and (3) any caveats the issuing
// server attaches. Implementations are free to grant a narrower set than the
// caller requested; callers should check the returned macaroon for the
// actions actually permitted.
//
// The action namespace uses dotted prefixes. Future protocol versions may add
// values; clients tolerate unknown action values gracefully.
type SessionAction string

const (
	// SessionActionJourneyRead permits reading a specific journey's data.
	SessionActionJourneyRead SessionAction = "journey.read"
	// SessionActionJourneyWrite permits mutating a specific journey's data.
	SessionActionJourneyWrite SessionAction = "journey.write"
	// SessionActionTelemetryWrite permits submitting position samples to a
	// specific journey.
	SessionActionTelemetryWrite SessionAction = "telemetry.write"
	// SessionActionMediaUpload permits uploading shared media to a specific
	// journey.
	SessionActionMediaUpload SessionAction = "media.upload"
	// SessionActionInviteCreate permits creating journey invites within the
	// caller's invite-generation permissions.
	SessionActionInviteCreate SessionAction = "invite.create"
)

// Valid reports whether the action is a known OpenCaravan value.
func (a SessionAction) Valid() bool {
	switch a {
	case SessionActionJourneyRead,
		SessionActionJourneyWrite,
		SessionActionTelemetryWrite,
		SessionActionMediaUpload,
		SessionActionInviteCreate:
		return true
	default:
		return false
	}
}

// SessionRequest is the wire format a client app sends, over an
// mTLS-authenticated transport, to request a session macaroon.
//
// JourneyID is optional; a session without a journey ID grants actions that
// do not require a journey context (such as creating a journey or reading the
// caller's own profile). LifetimeSeconds is a hint — the server may issue a
// macaroon with a shorter lifetime than requested.
type SessionRequest struct {
	Type            string          `json:"type"`
	Version         int             `json:"version"`
	JourneyID       *UUID           `json:"journey_id,omitempty"`
	Actions         []SessionAction `json:"actions"`
	LifetimeSeconds int             `json:"lifetime_seconds"`
}

// Validate reports whether the request has the required type, version, action
// list, and a non-negative lifetime hint.
func (r SessionRequest) Validate() error {
	if r.Type != SessionRequestType {
		return fmt.Errorf("type must be %q", SessionRequestType)
	}
	if r.Version != SessionVersion {
		return fmt.Errorf("version must be %d", SessionVersion)
	}
	if r.JourneyID != nil && !r.JourneyID.Valid() {
		return errors.New("journey_id must be a valid UUID")
	}
	if len(r.Actions) == 0 {
		return errors.New("actions must contain at least one entry")
	}
	for i, action := range r.Actions {
		if !action.Valid() {
			return fmt.Errorf("actions[%d] must be a known OpenCaravan value", i)
		}
	}
	if r.LifetimeSeconds < 0 {
		return errors.New("lifetime_seconds must be non-negative")
	}
	return nil
}

// SessionResponse is the server's reply to a SessionRequest, carrying the
// issued macaroon and its expiration time.
//
// Macaroon carries the binary macaroon serialization defined by the macaroon
// spec (https://research.google/pubs/macaroons-cookies-with-contextual-caveats-for-decentralized-authorization/)
// encoded as unpadded base64url so it travels safely in HTTP headers, JSON
// payloads, and URL query strings without escaping. OpenCaravan defines only
// the caveat predicate namespace used inside the macaroon (see CaveatKind and
// the CaveatX builders); the macaroon binary format itself is out of scope
// for this protocol package.
//
// The ExpirationTime field duplicates the time<T caveat inside the macaroon
// so clients can show expiry without having to parse the macaroon body.
type SessionResponse struct {
	Type           string    `json:"type"`
	Version        int       `json:"version"`
	Macaroon       string    `json:"macaroon"`
	ExpirationTime time.Time `json:"expiration_time"`
}

// Validate reports whether the response has the required type, version,
// macaroon payload, and expiration time.
func (r SessionResponse) Validate() error {
	if r.Type != SessionResponseType {
		return fmt.Errorf("type must be %q", SessionResponseType)
	}
	if r.Version != SessionVersion {
		return fmt.Errorf("version must be %d", SessionVersion)
	}
	if strings.TrimSpace(r.Macaroon) == "" {
		return errors.New("macaroon must be set")
	}
	if _, err := base64.RawURLEncoding.DecodeString(r.Macaroon); err != nil {
		return fmt.Errorf("macaroon must be unpadded base64url: %w", err)
	}
	if r.ExpirationTime.IsZero() {
		return errors.New("expiration_time must be set")
	}
	return nil
}

// NewSessionRequest returns a SessionRequest with the current type and
// version fields populated.
func NewSessionRequest(actions []SessionAction, lifetimeSeconds int) SessionRequest {
	return SessionRequest{
		Type:            SessionRequestType,
		Version:         SessionVersion,
		Actions:         actions,
		LifetimeSeconds: lifetimeSeconds,
	}
}

// NewSessionResponse returns a SessionResponse with the current type and
// version fields populated.
func NewSessionResponse(macaroon string, expirationTime time.Time) SessionResponse {
	return SessionResponse{
		Type:           SessionResponseType,
		Version:        SessionVersion,
		Macaroon:       macaroon,
		ExpirationTime: expirationTime,
	}
}
