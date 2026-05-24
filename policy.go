package opencaravan

import (
	"errors"
	"fmt"
	"time"
)

// ProtocolVersion is the current draft OpenCaravan protocol version.
const ProtocolVersion = "0.1-draft"

// RegistrationMode describes how a server allows new accounts or devices to
// enroll.
type RegistrationMode string

const (
	// RegistrationClosed means the server does not accept self-service
	// enrollment.
	RegistrationClosed RegistrationMode = "closed"
	// RegistrationInvite means enrollment requires an invite.
	RegistrationInvite RegistrationMode = "invite"
	// RegistrationOpen means the server accepts public enrollment.
	RegistrationOpen RegistrationMode = "open"
)

// Valid reports whether the registration mode is a known OpenCaravan value.
func (m RegistrationMode) Valid() bool {
	switch m {
	case RegistrationClosed, RegistrationInvite, RegistrationOpen:
		return true
	default:
		return false
	}
}

// JourneyRetentionMode describes how long a server will keep journey data after
// the live journey lifecycle ends.
type JourneyRetentionMode string

const (
	// JourneyRetentionEphemeral means journey data is hard-purged at or soon
	// after the journey closes or expires.
	JourneyRetentionEphemeral JourneyRetentionMode = "ephemeral"
	// JourneyRetentionDownloadWindow means journey data remains available for
	// participant export until a bounded download window closes.
	JourneyRetentionDownloadWindow JourneyRetentionMode = "download_window"
	// JourneyRetentionForever means journey data is retained without a scheduled
	// purge deadline.
	JourneyRetentionForever JourneyRetentionMode = "forever"
)

// Valid reports whether the journey retention mode is a known OpenCaravan
// value.
func (m JourneyRetentionMode) Valid() bool {
	switch m {
	case JourneyRetentionEphemeral, JourneyRetentionDownloadWindow, JourneyRetentionForever:
		return true
	default:
		return false
	}
}

// ServerPolicy advertises a server's public OpenCaravan capability envelope.
type ServerPolicy struct {
	ProtocolVersion        string                 `json:"protocol_version"`
	ServerURL              string                 `json:"server_url"`
	DisplayName            string                 `json:"display_name"`
	RegistrationMode       RegistrationMode       `json:"registration_mode"`
	JourneyPolicies        []JourneyPolicyProfile `json:"journey_policies"`
	DefaultJourneyPolicyID string                 `json:"default_journey_policy_id,omitempty"`
	PrivacyURL             string                 `json:"privacy_url,omitempty"`
	TermsURL               string                 `json:"terms_url,omitempty"`
	Metadata               map[string]any         `json:"metadata,omitempty"`
}

// JourneyPolicyProfile is one journey lifecycle policy a server offers during
// journey creation.
//
// The profile is the advertised menu item. A created journey stores a
// JourneyPolicy snapshot with concrete timestamps resolved by the server.
type JourneyPolicyProfile struct {
	ID                string               `json:"id"`
	DisplayName       string               `json:"display_name"`
	Description       string               `json:"description,omitempty"`
	RetentionMode     JourneyRetentionMode `json:"retention_mode"`
	MaxActiveLifetime string               `json:"max_active_lifetime,omitempty"`
	MaxDownloadWindow string               `json:"max_download_window,omitempty"`
	ExportSupported   bool                 `json:"export_supported"`
	MediaAllowed      bool                 `json:"media_allowed"`
	Metadata          map[string]any       `json:"metadata,omitempty"`
}

// JourneyPolicy is the concrete lifecycle policy snapshot a participant accepts
// for one journey.
//
// PolicyID names the server-advertised profile selected at creation time.
// PolicyHash fingerprints the resolved snapshot. PurgeAt is a hard-deletion
// deadline, not a soft-delete marker.
type JourneyPolicy struct {
	PolicyID        string               `json:"policy_id"`
	PolicyHash      string               `json:"policy_hash"`
	DisplayName     string               `json:"display_name,omitempty"`
	RetentionMode   JourneyRetentionMode `json:"retention_mode"`
	ActiveExpiresAt *time.Time           `json:"active_expires_at,omitempty"`
	DownloadUntil   *time.Time           `json:"download_until,omitempty"`
	PurgeAt         *time.Time           `json:"purge_at,omitempty"`
	RetainForever   bool                 `json:"retain_forever"`
	ExportSupported bool                 `json:"export_supported"`
	MediaAllowed    bool                 `json:"media_allowed"`
}

// Validate reports whether policy advertises a coherent server policy menu.
func (policy ServerPolicy) Validate() error {
	if policy.ProtocolVersion == "" {
		return errors.New("protocol_version must be set")
	}
	if policy.ServerURL == "" {
		return errors.New("server_url must be set")
	}
	if policy.DisplayName == "" {
		return errors.New("display_name must be set")
	}
	if !policy.RegistrationMode.Valid() {
		return errors.New("registration_mode must be a known OpenCaravan value")
	}
	if len(policy.JourneyPolicies) == 0 {
		return errors.New("journey_policies must contain at least one policy profile")
	}

	profileIDs := make(map[string]struct{}, len(policy.JourneyPolicies))
	for i, profile := range policy.JourneyPolicies {
		if err := profile.Validate(); err != nil {
			return fmt.Errorf("journey_policies[%d]: %w", i, err)
		}
		if _, exists := profileIDs[profile.ID]; exists {
			return fmt.Errorf("journey_policies[%d]: duplicate policy id %q", i, profile.ID)
		}
		profileIDs[profile.ID] = struct{}{}
	}
	if policy.DefaultJourneyPolicyID != "" {
		if _, exists := profileIDs[policy.DefaultJourneyPolicyID]; !exists {
			return errors.New("default_journey_policy_id must match a journey policy profile")
		}
	}
	return nil
}

// Validate reports whether profile is a coherent server-offered journey policy.
func (profile JourneyPolicyProfile) Validate() error {
	if profile.ID == "" {
		return errors.New("id must be set")
	}
	if profile.DisplayName == "" {
		return errors.New("display_name must be set")
	}
	if !profile.RetentionMode.Valid() {
		return errors.New("retention_mode must be a known OpenCaravan value")
	}
	switch profile.RetentionMode {
	case JourneyRetentionEphemeral:
		if profile.MaxDownloadWindow != "" {
			return errors.New("ephemeral policy profile cannot advertise a download window")
		}
		if profile.ExportSupported {
			return errors.New("ephemeral policy profile cannot support retained export")
		}
	case JourneyRetentionDownloadWindow:
		if !profile.ExportSupported {
			return errors.New("download-window policy profile must support export")
		}
	case JourneyRetentionForever:
		if !profile.ExportSupported {
			return errors.New("forever policy profile must support export")
		}
	}
	return nil
}

// Validate reports whether policy is a coherent journey lifecycle snapshot.
func (policy JourneyPolicy) Validate() error {
	if policy.PolicyID == "" {
		return errors.New("policy_id must be set")
	}
	if policy.PolicyHash == "" {
		return errors.New("policy_hash must be set")
	}
	if !policy.RetentionMode.Valid() {
		return errors.New("retention_mode must be a known OpenCaravan value")
	}
	if policy.ActiveExpiresAt != nil && policy.ActiveExpiresAt.IsZero() {
		return errors.New("active_expires_at must be a non-zero time")
	}
	if policy.DownloadUntil != nil && policy.DownloadUntil.IsZero() {
		return errors.New("download_until must be a non-zero time")
	}
	if policy.PurgeAt != nil && policy.PurgeAt.IsZero() {
		return errors.New("purge_at must be a non-zero time")
	}
	if policy.DownloadUntil != nil && policy.PurgeAt != nil && policy.PurgeAt.Before(*policy.DownloadUntil) {
		return errors.New("purge_at must not be before download_until")
	}

	switch policy.RetentionMode {
	case JourneyRetentionEphemeral:
		if policy.RetainForever {
			return errors.New("ephemeral journey policy cannot retain forever")
		}
		if policy.DownloadUntil != nil {
			return errors.New("ephemeral journey policy cannot set download_until")
		}
		if policy.PurgeAt == nil {
			return errors.New("ephemeral journey policy must set purge_at")
		}
		if policy.ExportSupported {
			return errors.New("ephemeral journey policy cannot support retained export")
		}
	case JourneyRetentionDownloadWindow:
		if policy.RetainForever {
			return errors.New("download-window journey policy cannot retain forever")
		}
		if policy.DownloadUntil == nil {
			return errors.New("download-window journey policy must set download_until")
		}
		if policy.PurgeAt == nil {
			return errors.New("download-window journey policy must set purge_at")
		}
		if !policy.ExportSupported {
			return errors.New("download-window journey policy must support export")
		}
	case JourneyRetentionForever:
		if !policy.RetainForever {
			return errors.New("forever journey policy must set retain_forever")
		}
		if policy.PurgeAt != nil {
			return errors.New("forever journey policy cannot set purge_at")
		}
		if !policy.ExportSupported {
			return errors.New("forever journey policy must support export")
		}
	}

	return nil
}
