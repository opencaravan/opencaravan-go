package opencaravan

import "time"

// JourneyVisibility describes who can discover or join a journey.
type JourneyVisibility string

const (
	// JourneyPrivate means the journey is visible only to known participants.
	JourneyPrivate JourneyVisibility = "private"
	// JourneyInvite means the journey can be joined through an invite.
	JourneyInvite JourneyVisibility = "invite"
	// JourneyServer means the journey can be discovered by authenticated users
	// on the same server.
	JourneyServer JourneyVisibility = "server"
	// JourneyPublic means the journey is publicly discoverable.
	JourneyPublic JourneyVisibility = "public"
)

// JourneyState describes the lifecycle state of a journey.
type JourneyState string

const (
	// JourneyPlanned means the journey exists but is not active.
	JourneyPlanned JourneyState = "planned"
	// JourneyActive means the journey is accepting live participant updates.
	JourneyActive JourneyState = "active"
	// JourneyClosed means the journey ended normally.
	JourneyClosed JourneyState = "closed"
	// JourneyExpired means the journey ended because its server-side lifetime
	// expired.
	JourneyExpired JourneyState = "expired"
	// JourneyDeleted means the journey is no longer available.
	JourneyDeleted JourneyState = "deleted"
)

// Journey describes a group drive shared through OpenCaravan.
type Journey struct {
	ID              string            `json:"id"`
	OriginServerURL string            `json:"origin_server_url"`
	Title           string            `json:"title"`
	Description     string            `json:"description,omitempty"`
	Visibility      JourneyVisibility `json:"visibility"`
	State           JourneyState      `json:"state"`
	PolicyHash      string            `json:"policy_hash"`
	CreatedAt       time.Time         `json:"created_at"`
	StartsAt        *time.Time        `json:"starts_at,omitempty"`
	StartedAt       *time.Time        `json:"started_at,omitempty"`
	ClosedAt        *time.Time        `json:"closed_at,omitempty"`
}

// ParticipantRole describes a participant's role in a journey.
type ParticipantRole string

const (
	// ParticipantHost means the participant coordinates the journey.
	ParticipantHost ParticipantRole = "host"
	// ParticipantDriver means the participant is driving a vehicle.
	ParticipantDriver ParticipantRole = "driver"
	// ParticipantPassenger means the participant is riding in a vehicle.
	ParticipantPassenger ParticipantRole = "passenger"
	// ParticipantObserver means the participant can observe but is not sharing
	// vehicle movement.
	ParticipantObserver ParticipantRole = "observer"
)

// ParticipantState describes whether a participant is currently part of a
// journey.
type ParticipantState string

const (
	// ParticipantInvited means the participant has not joined yet.
	ParticipantInvited ParticipantState = "invited"
	// ParticipantJoined means the participant has joined the journey.
	ParticipantJoined ParticipantState = "joined"
	// ParticipantLeft means the participant left voluntarily.
	ParticipantLeft ParticipantState = "left"
	// ParticipantRemoved means the participant was removed from the journey.
	ParticipantRemoved ParticipantState = "removed"
)

// SharingState describes whether a participant is sharing live position data.
type SharingState string

const (
	// SharingOff means the participant is not sharing location.
	SharingOff SharingState = "off"
	// SharingLive means the participant is sharing live location.
	SharingLive SharingState = "live"
	// SharingPaused means location sharing is temporarily paused.
	SharingPaused SharingState = "paused"
)

// Participant describes a person or client presence inside a journey.
type Participant struct {
	ID           string           `json:"id"`
	JourneyID    string           `json:"journey_id"`
	DisplayName  string           `json:"display_name"`
	Role         ParticipantRole  `json:"role"`
	State        ParticipantState `json:"state"`
	SharingState SharingState     `json:"sharing_state"`
	PolicyHash   string           `json:"policy_hash"`
	JoinedAt     *time.Time       `json:"joined_at,omitempty"`
	LastSeenAt   *time.Time       `json:"last_seen_at,omitempty"`
	LeftAt       *time.Time       `json:"left_at,omitempty"`
}

// Vehicle describes the user-visible vehicle identity attached to a journey
// participant.
type Vehicle struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Make        string `json:"make,omitempty"`
	Model       string `json:"model,omitempty"`
	ModelYear   int    `json:"model_year,omitempty"`
	Color       string `json:"color,omitempty"`
}
