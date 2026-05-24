package opencaravan

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"
)

const journeyInviteTokenBytes = 32

// JourneyInviteType is the canonical type value for journey invites.
const JourneyInviteType = "opencaravan.journey_invite"

// JourneyInviteVersion is the current journey invite object version.
const JourneyInviteVersion = 1

// JourneyInviteUseMode describes whether an invite token may be redeemed once
// or more than once.
type JourneyInviteUseMode string

const (
	// JourneyInviteSingleUse means the token can admit at most one prospective
	// participant.
	JourneyInviteSingleUse JourneyInviteUseMode = "single_use"
	// JourneyInviteMultiUse means the token can admit more than one prospective
	// participant, optionally capped by MaxUses.
	JourneyInviteMultiUse JourneyInviteUseMode = "multi_use"
)

// Valid reports whether the use mode is a known OpenCaravan value.
func (m JourneyInviteUseMode) Valid() bool {
	switch m {
	case JourneyInviteSingleUse, JourneyInviteMultiUse:
		return true
	default:
		return false
	}
}

// JourneyInviteAudience describes the expected sharing pattern for an invite.
//
// Audience helps clients choose good presentation and warning language. The
// server still enforces actual redemption behavior through the token use mode
// and any server-side policy tied to the invite.
type JourneyInviteAudience string

const (
	// JourneyInviteGroupAudience means the invite is intended for a group chat,
	// web forum, or other place where multiple people may redeem it.
	JourneyInviteGroupAudience JourneyInviteAudience = "group"
	// JourneyInviteIndividualAudience means the invite is intended for one
	// prospective participant through a private message or direct share.
	JourneyInviteIndividualAudience JourneyInviteAudience = "individual"
)

// Valid reports whether the audience is a known OpenCaravan value.
func (a JourneyInviteAudience) Valid() bool {
	switch a {
	case JourneyInviteGroupAudience, JourneyInviteIndividualAudience:
		return true
	default:
		return false
	}
}

// JourneyInvite is a portable, integrity-protected invitation to a private
// journey.
//
// Apps may encode the same invite as a universal link, QR payload, AirDrop
// payload, email, chat message, or other platform-native share surface. The
// token is the secret capability. Integrity records how the issuing server made
// the rest of the object tamper-evident.
type JourneyInvite struct {
	Type                          string                     `json:"type"`
	Version                       int                        `json:"version"`
	ID                            UUID                       `json:"id"`
	ServerURL                     string                     `json:"server_url"`
	JourneyID                     UUID                       `json:"journey_id"`
	Token                         JourneyInviteToken         `json:"token"`
	Audience                      JourneyInviteAudience      `json:"audience"`
	CreatedByJourneyParticipantID UUID                       `json:"created_by_journey_participant_id"`
	CreatedAt                     time.Time                  `json:"created_at"`
	PolicyHash                    string                     `json:"policy_hash"`
	DisplayName                   string                     `json:"display_name,omitempty"`
	Links                         *JourneyInviteLinks        `json:"links,omitempty"`
	Presentation                  *JourneyInvitePresentation `json:"presentation,omitempty"`
	Integrity                     *JourneyInviteIntegrity    `json:"integrity"`
}

// JourneyInviteToken is the server-issued secret capability carried by a
// journey invite.
type JourneyInviteToken struct {
	Value     string               `json:"value"`
	UseMode   JourneyInviteUseMode `json:"use_mode"`
	MaxUses   int                  `json:"max_uses,omitempty"`
	ExpiresAt time.Time            `json:"expires_at"`
}

// JourneyInviteLinks describes the URL forms an app or server can use to
// process an invite.
type JourneyInviteLinks struct {
	WebURL string `json:"web_url,omitempty"`
	AppURL string `json:"app_url,omitempty"`
}

// JourneyInvitePresentation contains display hints for rich platform-native
// invite sharing.
type JourneyInvitePresentation struct {
	Title    string         `json:"title,omitempty"`
	Summary  string         `json:"summary,omitempty"`
	ImageURL string         `json:"image_url,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// JourneyInviteIntegrity describes the signature or message authentication
// data that makes a journey invite tamper-evident.
//
// Signature covers the canonical invite object excluding the Integrity field.
// KeyID identifies the issuer key a client can use to verify the signature.
type JourneyInviteIntegrity struct {
	Algorithm string `json:"algorithm"`
	KeyID     string `json:"key_id,omitempty"`
	Signature string `json:"signature"`
}

// NewJourneyInviteToken returns a cryptographically random invite token.
//
// The token value contains 256 bits of randomness encoded as unpadded base64url
// text so it can travel safely in URLs, QR codes, JSON, and platform share
// payloads.
func NewJourneyInviteToken(useMode JourneyInviteUseMode, expiresAt time.Time) (JourneyInviteToken, error) {
	if !useMode.Valid() {
		return JourneyInviteToken{}, errors.New("invite token use mode must be a known OpenCaravan value")
	}
	if expiresAt.IsZero() {
		return JourneyInviteToken{}, errors.New("invite token expires_at must be set")
	}

	b := make([]byte, journeyInviteTokenBytes)
	if _, err := rand.Read(b); err != nil {
		return JourneyInviteToken{}, fmt.Errorf("read random invite token bytes: %w", err)
	}

	token := JourneyInviteToken{
		Value:     base64.RawURLEncoding.EncodeToString(b),
		UseMode:   useMode,
		ExpiresAt: expiresAt,
	}
	if useMode == JourneyInviteSingleUse {
		token.MaxUses = 1
	}
	return token, nil
}

// NewJourneyInvite returns a journey invite with the current type and version
// fields populated.
func NewJourneyInvite(serverURL string, journeyID UUID, token JourneyInviteToken) JourneyInvite {
	return JourneyInvite{
		Type:      JourneyInviteType,
		Version:   JourneyInviteVersion,
		ServerURL: serverURL,
		JourneyID: journeyID,
		Token:     token,
	}
}

// Validate reports whether invite contains the required identity, capability,
// issuer, policy fingerprint, and integrity fields.
func (invite JourneyInvite) Validate() error {
	if invite.Type != JourneyInviteType {
		return fmt.Errorf("type must be %q", JourneyInviteType)
	}
	if invite.Version != JourneyInviteVersion {
		return fmt.Errorf("version must be %d", JourneyInviteVersion)
	}
	if !invite.ID.Valid() {
		return errors.New("invite id must be a valid UUID")
	}
	if invite.ServerURL == "" {
		return errors.New("server_url must be set")
	}
	if !invite.JourneyID.Valid() {
		return errors.New("journey_id must be a valid UUID")
	}
	if err := invite.Token.Validate(); err != nil {
		return fmt.Errorf("token: %w", err)
	}
	if !invite.Audience.Valid() {
		return errors.New("audience must be a known OpenCaravan value")
	}
	if !invite.CreatedByJourneyParticipantID.Valid() {
		return errors.New("created_by_journey_participant_id must be a valid UUID")
	}
	if invite.CreatedAt.IsZero() {
		return errors.New("created_at must be set")
	}
	if invite.PolicyHash == "" {
		return errors.New("policy_hash must be set")
	}
	if invite.Integrity == nil {
		return errors.New("integrity must be set")
	}
	if err := invite.Integrity.Validate(); err != nil {
		return fmt.Errorf("integrity: %w", err)
	}
	return nil
}

// Validate reports whether token contains a secret value, known use mode,
// bounded use count semantics, and expiration.
func (token JourneyInviteToken) Validate() error {
	if token.Value == "" {
		return errors.New("invite token value must be set")
	}
	tokenBytes, err := base64.RawURLEncoding.DecodeString(token.Value)
	if err != nil {
		return fmt.Errorf("invite token value must be unpadded base64url: %w", err)
	}
	if len(tokenBytes) != journeyInviteTokenBytes {
		return fmt.Errorf("invite token must contain %d random bytes", journeyInviteTokenBytes)
	}
	if !token.UseMode.Valid() {
		return errors.New("invite token use mode must be a known OpenCaravan value")
	}
	if token.MaxUses < 0 {
		return errors.New("invite token max_uses must be non-negative")
	}
	switch token.UseMode {
	case JourneyInviteSingleUse:
		if token.MaxUses > 1 {
			return errors.New("single-use invite token max_uses must be 0 or 1")
		}
	case JourneyInviteMultiUse:
		if token.MaxUses == 1 {
			return errors.New("multi-use invite token max_uses must be 0 or greater than 1")
		}
	}
	if token.ExpiresAt.IsZero() {
		return errors.New("invite token expires_at must be set")
	}
	return nil
}

// Validate reports whether integrity contains the fields needed to verify a
// signed invite object.
func (integrity JourneyInviteIntegrity) Validate() error {
	if integrity.Algorithm == "" {
		return errors.New("algorithm must be set")
	}
	if integrity.Signature == "" {
		return errors.New("signature must be set")
	}
	return nil
}
