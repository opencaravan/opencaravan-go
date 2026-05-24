package opencaravan

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

// RetentionMode describes whether journey data is discarded or retained after
// the journey ends.
type RetentionMode string

const (
	// RetentionEphemeral means journey telemetry is discarded when the journey
	// lifecycle closes.
	RetentionEphemeral RetentionMode = "ephemeral"
	// RetentionRetained means journey telemetry may be retained under the
	// journey's advertised policy.
	RetentionRetained RetentionMode = "retained"
)

// Valid reports whether the retention mode is a known OpenCaravan value.
func (m RetentionMode) Valid() bool {
	switch m {
	case RetentionEphemeral, RetentionRetained:
		return true
	default:
		return false
	}
}

// ServerPolicy advertises a server's public OpenCaravan capability envelope.
type ServerPolicy struct {
	ProtocolVersion  string                `json:"protocol_version"`
	ServerURL        string                `json:"server_url"`
	DisplayName      string                `json:"display_name"`
	RegistrationMode RegistrationMode      `json:"registration_mode"`
	Retention        RetentionCapabilities `json:"retention"`
	PrivacyURL       string                `json:"privacy_url,omitempty"`
	TermsURL         string                `json:"terms_url,omitempty"`
	Metadata         map[string]any        `json:"metadata,omitempty"`
}

// RetentionCapabilities advertises the journey retention modes and limits a
// server can support.
type RetentionCapabilities struct {
	Modes                []RetentionMode `json:"modes"`
	DefaultMode          RetentionMode   `json:"default_mode"`
	MaxLocationRetention string          `json:"max_location_retention,omitempty"`
	MaxMetadataRetention string          `json:"max_metadata_retention,omitempty"`
	ExportSupported      bool            `json:"export_supported"`
	DeleteSupported      bool            `json:"delete_supported"`
}

// JourneyPolicy is the policy snapshot a participant accepts for one journey.
type JourneyPolicy struct {
	PolicyHash string          `json:"policy_hash"`
	Retention  RetentionPolicy `json:"retention"`
	Metadata   map[string]any  `json:"metadata,omitempty"`
}

// RetentionPolicy describes how a journey will handle retained participant and
// telemetry data.
type RetentionPolicy struct {
	Mode                RetentionMode `json:"mode"`
	LocationSamples     string        `json:"location_samples"`
	ParticipantPresence string        `json:"participant_presence"`
	JourneyMetadata     string        `json:"journey_metadata"`
}
