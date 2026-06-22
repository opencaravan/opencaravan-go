package opencaravan

import "errors"

// ProtocolVersion is the current draft OpenCaravan protocol version.
//
// 0.2-draft introduced the content-addressed [BlobRef] type, refactored
// [Vehicle] into a metadata-only signed bundle (moving authorization
// fields to [VehicleACL] exclusively), and replaced the URL-based image
// references on [Vehicle] / [GarageVehicle] with hash-based [BlobRef]
// references.
const ProtocolVersion = "0.2-draft"

// RegistrationMode describes whether a server is currently accepting
// invite-backed user registration.
type RegistrationMode string

const (
	// RegistrationClosed means the server does not accept new user
	// registrations.
	RegistrationClosed RegistrationMode = "closed"
	// RegistrationInvite means user registration requires a server or journey
	// invite with registration scope.
	RegistrationInvite RegistrationMode = "invite"
)

// Valid reports whether the registration mode is a known OpenCaravan value.
func (m RegistrationMode) Valid() bool {
	switch m {
	case RegistrationClosed, RegistrationInvite:
		return true
	default:
		return false
	}
}

// ServerPolicy advertises a server's public OpenCaravan capability envelope.
//
// Journey-specific retention and feature choices are represented directly on
// Journey. ServerPolicy describes the server identity, invite-governed
// enrollment posture, and operator-facing policy documents that users should
// see before joining.
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
