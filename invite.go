package opencaravan

import "time"

// JourneyInvitePayloadType is the canonical type value for journey invite
// payloads.
const JourneyInvitePayloadType = "opencaravan.journey_invite"

// JourneyInvitePayloadVersion is the current journey invite payload version.
const JourneyInvitePayloadVersion = 1

// JourneyInvitePayload is a portable invite that can be shared as a link, QR
// code, AirDrop payload, email, or chat message.
type JourneyInvitePayload struct {
	Type        string    `json:"type"`
	Version     int       `json:"version"`
	ServerURL   string    `json:"server_url"`
	JourneyID   UUID      `json:"journey_id"`
	InviteToken string    `json:"invite_token"`
	ExpiresAt   time.Time `json:"expires_at"`
	PolicyHash  string    `json:"policy_hash"`
	DisplayName string    `json:"display_name,omitempty"`
}

// NewJourneyInvitePayload returns a journey invite payload with the current
// type and version fields populated.
func NewJourneyInvitePayload(serverURL string, journeyID UUID, inviteToken string, expiresAt time.Time) JourneyInvitePayload {
	return JourneyInvitePayload{
		Type:        JourneyInvitePayloadType,
		Version:     JourneyInvitePayloadVersion,
		ServerURL:   serverURL,
		JourneyID:   journeyID,
		InviteToken: inviteToken,
		ExpiresAt:   expiresAt,
	}
}
