package opencaravan

import "errors"

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

// ServerPolicy advertises a server's public OpenCaravan capability envelope.
//
// Journey-specific retention and feature choices are represented directly on
// Journey. ServerPolicy describes the server identity, enrollment posture, and
// operator-facing policy documents that users should see before joining.
type ServerPolicy struct {
	ProtocolVersion  string           `json:"protocol_version"`
	ServerURL        string           `json:"server_url"`
	DisplayName      string           `json:"display_name"`
	RegistrationMode RegistrationMode `json:"registration_mode"`
	PrivacyURL       string           `json:"privacy_url,omitempty"`
	TermsURL         string           `json:"terms_url,omitempty"`
	Metadata         map[string]any   `json:"metadata,omitempty"`
}

// Validate reports whether policy advertises the required server identity and
// enrollment fields.
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
	return nil
}
