package opencaravan

import (
	"errors"
	"fmt"
	"time"
)

// JourneyInviteType is the canonical type value for journey invites.
const JourneyInviteType = "opencaravan.journey_invite"

// JourneyInviteVersion is the current journey invite object version.
const JourneyInviteVersion = 1

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
// describe the scopes a caller may request from a future GenerateInvite
// operation, plus optional caps the server may enforce.
type InviteGenerationPermissions struct {
	Scopes []InviteScope `json:"scopes,omitempty"`
	// MaxRedemptionsPerInvite limits how many successful redemptions a
	// generated invite may allow. Zero means the server has not set a cap.
	MaxRedemptionsPerInvite int `json:"max_redemptions_per_invite,omitempty"`
	MaxLifetimeDays         int `json:"max_lifetime_days,omitempty"`
}

// Validate reports whether permissions contain valid invite scopes and
// non-negative caps.
func (permissions InviteGenerationPermissions) Validate() error {
	if permissions.MaxRedemptionsPerInvite < 0 {
		return errors.New("max_redemptions_per_invite must be non-negative")
	}
	if permissions.MaxLifetimeDays < 0 {
		return errors.New("max_lifetime_days must be non-negative")
	}

	empty := len(permissions.Scopes) == 0 &&
		permissions.MaxRedemptionsPerInvite == 0 && permissions.MaxLifetimeDays == 0
	if empty {
		return nil
	}
	if len(permissions.Scopes) == 0 {
		return errors.New("scopes must contain at least one invite scope")
	}
	for i, scope := range permissions.Scopes {
		if !scope.Valid() {
			return fmt.Errorf("scopes[%d] must be a known OpenCaravan value", i)
		}
	}
	return nil
}

// JourneyInvite is a portable, integrity-protected invitation to a private
// journey.
//
// Apps may encode the same invite as a universal link, QR payload, AirDrop
// payload, email, chat message, or other platform-native share surface. The
// token is the secret capability. Integrity records how the issuing server made
// the rest of the object tamper-evident.
type JourneyInvite struct {
	Type      string      `json:"type"`
	Version   int         `json:"version"`
	ID        UUID        `json:"id"`
	ServerURL string      `json:"server_url"`
	JourneyID UUID        `json:"journey_id"`
	Token     InviteToken `json:"token"`
	// MaxRedemptions is the maximum number of successful token redemptions.
	// Zero means the issuing server has not capped redemptions.
	MaxRedemptions                int                        `json:"max_redemptions"`
	CreatedByJourneyParticipantID UUID                       `json:"created_by_journey_participant_id"`
	CreationTime                  time.Time                  `json:"creation_time"`
	PolicyHash                    string                     `json:"policy_hash"`
	DisplayName                   string                     `json:"display_name,omitempty"`
	Links                         *JourneyInviteLinks        `json:"links,omitempty"`
	Presentation                  *JourneyInvitePresentation `json:"presentation,omitempty"`
	Integrity                     *Integrity                 `json:"integrity"`
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
	Title   string `json:"title,omitempty"`
	Summary string `json:"summary,omitempty"`
	// AvatarImage is the compact image clients can use in invite previews.
	AvatarImage *ImageResourceRef `json:"avatar_image,omitempty"`
	// BannerImage is the wide image clients can use in richer invite previews.
	BannerImage *ImageResourceRef `json:"banner_image,omitempty"`
	Metadata    map[string]any    `json:"metadata,omitempty"`
}

// NewJourneyInvite returns a journey invite with the current type and version
// fields populated.
//
// maxRedemptions is copied to MaxRedemptions. Zero means the issuing server has
// not capped redemptions.
func NewJourneyInvite(serverURL string, journeyID UUID, token InviteToken, maxRedemptions int) JourneyInvite {
	return JourneyInvite{
		Type:           JourneyInviteType,
		Version:        JourneyInviteVersion,
		ServerURL:      serverURL,
		JourneyID:      journeyID,
		Token:          token,
		MaxRedemptions: maxRedemptions,
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
	if invite.MaxRedemptions < 0 {
		return errors.New("max_redemptions must be non-negative")
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
	if invite.Presentation != nil {
		if err := invite.Presentation.Validate(); err != nil {
			return fmt.Errorf("presentation: %w", err)
		}
	}
	if err := invite.Integrity.Validate(); err != nil {
		return fmt.Errorf("integrity: %w", err)
	}
	return nil
}

// Validate reports whether presentation contains valid optional image
// resources.
func (presentation JourneyInvitePresentation) Validate() error {
	if presentation.AvatarImage != nil {
		if err := presentation.AvatarImage.Validate(); err != nil {
			return fmt.Errorf("avatar_image: %w", err)
		}
	}
	if presentation.BannerImage != nil {
		if err := presentation.BannerImage.Validate(); err != nil {
			return fmt.Errorf("banner_image: %w", err)
		}
	}
	return nil
}
