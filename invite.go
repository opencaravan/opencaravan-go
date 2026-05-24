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

// InviteUseMode describes whether an invite token may be redeemed once or more
// than once.
type InviteUseMode string

const (
	// InviteSingleUse means the token can admit at most one prospective user or
	// participant.
	InviteSingleUse InviteUseMode = "single_use"
	// InviteMultiUse means the token can admit more than one prospective user
	// or participant, optionally capped by MaxUses.
	InviteMultiUse InviteUseMode = "multi_use"
)

// Valid reports whether the use mode is a known OpenCaravan value.
func (m InviteUseMode) Valid() bool {
	switch m {
	case InviteSingleUse, InviteMultiUse:
		return true
	default:
		return false
	}
}

// InviteScope describes what a generated invite is allowed to grant.
type InviteScope string

const (
	// InviteScopeJourney means the invite may grant access to a journey.
	InviteScopeJourney InviteScope = "journey"
	// InviteScopeServerRegistration means the invite may register a
	// server-scoped user.
	InviteScopeServerRegistration InviteScope = "server_registration"
)

// Valid reports whether scope is a known OpenCaravan invite scope.
func (scope InviteScope) Valid() bool {
	switch scope {
	case InviteScopeJourney, InviteScopeServerRegistration:
		return true
	default:
		return false
	}
}

// InviteGenerationPermissions describes the invite power a server has granted
// to a user or journey participant.
//
// Empty permissions grant no invite generation power. Non-empty permissions
// describe the scopes and use modes a caller may request from a future
// GenerateInvite operation, plus optional caps the server may enforce.
type InviteGenerationPermissions struct {
	Scopes           []InviteScope   `json:"scopes,omitempty"`
	UseModes         []InviteUseMode `json:"use_modes,omitempty"`
	MaxUsesPerInvite int             `json:"max_uses_per_invite,omitempty"`
	MaxLifetimeDays  int             `json:"max_lifetime_days,omitempty"`
}

// Validate reports whether permissions contain valid invite scopes, use modes,
// and non-negative caps.
func (permissions InviteGenerationPermissions) Validate() error {
	if permissions.MaxUsesPerInvite < 0 {
		return errors.New("max_uses_per_invite must be non-negative")
	}
	if permissions.MaxLifetimeDays < 0 {
		return errors.New("max_lifetime_days must be non-negative")
	}

	empty := len(permissions.Scopes) == 0 && len(permissions.UseModes) == 0 &&
		permissions.MaxUsesPerInvite == 0 && permissions.MaxLifetimeDays == 0
	if empty {
		return nil
	}
	if len(permissions.Scopes) == 0 {
		return errors.New("scopes must contain at least one invite scope")
	}
	if len(permissions.UseModes) == 0 {
		return errors.New("use_modes must contain at least one invite use mode")
	}
	for i, scope := range permissions.Scopes {
		if !scope.Valid() {
			return fmt.Errorf("scopes[%d] must be a known OpenCaravan value", i)
		}
	}
	for i, mode := range permissions.UseModes {
		if !mode.Valid() {
			return fmt.Errorf("use_modes[%d] must be a known OpenCaravan value", i)
		}
		if mode == InviteMultiUse && permissions.MaxUsesPerInvite == 1 {
			return errors.New("multi-use invite permissions max_uses_per_invite must be 0 or greater than 1")
		}
	}
	return nil
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
	CreationTime                  time.Time                  `json:"creation_time"`
	PolicyHash                    string                     `json:"policy_hash"`
	DisplayName                   string                     `json:"display_name,omitempty"`
	Links                         *JourneyInviteLinks        `json:"links,omitempty"`
	Presentation                  *JourneyInvitePresentation `json:"presentation,omitempty"`
	Integrity                     *JourneyInviteIntegrity    `json:"integrity"`
}

// JourneyInviteToken is the server-issued secret capability carried by a
// journey invite.
type JourneyInviteToken struct {
	Value          string        `json:"value"`
	UseMode        InviteUseMode `json:"use_mode"`
	MaxUses        int           `json:"max_uses,omitempty"`
	ExpirationTime time.Time     `json:"expiration_time"`
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
func NewJourneyInviteToken(useMode InviteUseMode, expirationTime time.Time) (JourneyInviteToken, error) {
	if !useMode.Valid() {
		return JourneyInviteToken{}, errors.New("invite token use mode must be a known OpenCaravan value")
	}
	if expirationTime.IsZero() {
		return JourneyInviteToken{}, errors.New("invite token expiration_time must be set")
	}

	b := make([]byte, journeyInviteTokenBytes)
	if _, err := rand.Read(b); err != nil {
		return JourneyInviteToken{}, fmt.Errorf("read random invite token bytes: %w", err)
	}

	token := JourneyInviteToken{
		Value:          base64.RawURLEncoding.EncodeToString(b),
		UseMode:        useMode,
		ExpirationTime: expirationTime,
	}
	if useMode == InviteSingleUse {
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
	if invite.CreationTime.IsZero() {
		return errors.New("creation_time must be set")
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
	case InviteSingleUse:
		if token.MaxUses > 1 {
			return errors.New("single-use invite token max_uses must be 0 or 1")
		}
	case InviteMultiUse:
		if token.MaxUses == 1 {
			return errors.New("multi-use invite token max_uses must be 0 or greater than 1")
		}
	}
	if token.ExpirationTime.IsZero() {
		return errors.New("invite token expiration_time must be set")
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
