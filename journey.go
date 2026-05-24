package opencaravan

import (
	"errors"
	"fmt"
	"time"
)

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
//
// Journey is the aggregate representation: APIs may return segments and media
// inline for small journeys, or page those collections separately for large
// retained journeys.
type Journey struct {
	ID              UUID              `json:"id"`
	OriginServerURL string            `json:"origin_server_url"`
	Title           string            `json:"title"`
	Description     string            `json:"description,omitempty"`
	Visibility      JourneyVisibility `json:"visibility"`
	State           JourneyState      `json:"state"`
	Policy          JourneyPolicy     `json:"policy"`
	Segments        []JourneySegment  `json:"segments,omitempty"`
	SharedMedia     []SharedMedia     `json:"shared_media,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	StartsAt        *time.Time        `json:"starts_at,omitempty"`
	StartedAt       *time.Time        `json:"started_at,omitempty"`
	ClosedAt        *time.Time        `json:"closed_at,omitempty"`
}

// HumanParticipant describes a person participating in OpenCaravan journeys.
//
// A human may run more than one OpenCaravan client app. Segment-level occupant
// records describe what the human is doing in a particular vehicle at a
// particular time.
type HumanParticipant struct {
	ID          UUID        `json:"id"`
	DisplayName string      `json:"display_name"`
	HomeServer  string      `json:"home_server,omitempty"`
	ClientApps  []ClientApp `json:"client_apps,omitempty"`
}

// ClientApp describes one OpenCaravan-capable app installation or session
// acting on behalf of a human participant.
type ClientApp struct {
	ID            UUID       `json:"id"`
	ParticipantID UUID       `json:"participant_id"`
	Name          string     `json:"name"`
	Version       string     `json:"version,omitempty"`
	Platform      string     `json:"platform,omitempty"`
	DeviceName    string     `json:"device_name,omitempty"`
	Capabilities  []string   `json:"capabilities,omitempty"`
	LastSeenAt    *time.Time `json:"last_seen_at,omitempty"`
}

// Vehicle describes a physical vehicle that can carry one or more participants
// during a journey segment.
type Vehicle struct {
	ID          UUID   `json:"id"`
	DisplayName string `json:"display_name"`
	Make        string `json:"make,omitempty"`
	Model       string `json:"model,omitempty"`
	ModelYear   int    `json:"model_year,omitempty"`
	Color       string `json:"color,omitempty"`
}

// SegmentState describes whether a journey segment is receiving or retaining
// tracklog data.
type SegmentState string

const (
	// SegmentPlanned means the segment has been defined but has not started.
	SegmentPlanned SegmentState = "planned"
	// SegmentActive means the segment is accepting participant position
	// samples.
	SegmentActive SegmentState = "active"
	// SegmentClosed means the segment ended normally and its tracklog is no
	// longer live.
	SegmentClosed SegmentState = "closed"
	// SegmentDiscarded means the segment should not retain tracklog data.
	SegmentDiscarded SegmentState = "discarded"
)

// JourneySegment describes a bounded portion of a journey.
//
// A segment contains the participating vehicles. Each segment vehicle contains
// its occupants and the tracklog samples that client apps submit for that
// vehicle during the segment.
type JourneySegment struct {
	ID        UUID             `json:"id"`
	JourneyID UUID             `json:"journey_id"`
	Name      string           `json:"name,omitempty"`
	State     SegmentState     `json:"state"`
	Vehicles  []SegmentVehicle `json:"vehicles,omitempty"`
	StartedAt time.Time        `json:"started_at"`
	EndedAt   *time.Time       `json:"ended_at,omitempty"`
}

// SegmentVehicle describes one vehicle's participation in a journey segment.
type SegmentVehicle struct {
	ID        UUID              `json:"id"`
	SegmentID UUID              `json:"segment_id"`
	VehicleID UUID              `json:"vehicle_id"`
	Occupants []VehicleOccupant `json:"occupants"`
	Tracklog  []PositionSample  `json:"tracklog,omitempty"`
}

// OccupantRole describes what a human participant is doing in a vehicle during
// a journey segment.
type OccupantRole string

const (
	// OccupantDriver means the participant is driving the vehicle.
	OccupantDriver OccupantRole = "driver"
	// OccupantNavigator means the participant is navigating or coordinating
	// from inside the vehicle.
	OccupantNavigator OccupantRole = "navigator"
	// OccupantRider means the participant is riding in the vehicle without a
	// driver or navigator role.
	OccupantRider OccupantRole = "rider"
)

// Valid reports whether the occupant role is a known OpenCaravan value.
func (r OccupantRole) Valid() bool {
	switch r {
	case OccupantDriver, OccupantNavigator, OccupantRider:
		return true
	default:
		return false
	}
}

// VehicleOccupant links a human participant and one or more client apps to a
// vehicle during a journey segment.
type VehicleOccupant struct {
	ParticipantID UUID         `json:"participant_id"`
	ClientAppIDs  []UUID       `json:"client_app_ids,omitempty"`
	Role          OccupantRole `json:"role"`
	JoinedAt      time.Time    `json:"joined_at"`
	LeftAt        *time.Time   `json:"left_at,omitempty"`
}

// Validate reports whether vehicle has the required segment, vehicle, occupant,
// and tracklog identity relationships.
func (vehicle SegmentVehicle) Validate() error {
	if !vehicle.ID.Valid() {
		return errors.New("segment vehicle id must be a valid UUID")
	}
	if !vehicle.SegmentID.Valid() {
		return errors.New("segment id must be a valid UUID")
	}
	if !vehicle.VehicleID.Valid() {
		return errors.New("vehicle id must be a valid UUID")
	}
	if len(vehicle.Occupants) == 0 {
		return errors.New("segment vehicle must contain at least one occupant")
	}

	participants := make(map[UUID]struct{}, len(vehicle.Occupants))
	for i, occupant := range vehicle.Occupants {
		if err := occupant.Validate(); err != nil {
			return fmt.Errorf("occupant %d: %w", i, err)
		}
		participants[occupant.ParticipantID] = struct{}{}
	}
	for i, sample := range vehicle.Tracklog {
		if err := sample.Validate(); err != nil {
			return fmt.Errorf("tracklog sample %d: %w", i, err)
		}
		if sample.SegmentVehicleID != vehicle.ID {
			return fmt.Errorf("tracklog sample %d: segment_vehicle_id does not match vehicle", i)
		}
		if _, ok := participants[sample.ParticipantID]; !ok {
			return fmt.Errorf("tracklog sample %d: participant_id is not an occupant", i)
		}
	}

	return nil
}

// Validate reports whether occupant contains a participant, valid role, and
// join time.
func (occupant VehicleOccupant) Validate() error {
	if !occupant.ParticipantID.Valid() {
		return errors.New("participant id must be a valid UUID")
	}
	if !occupant.Role.Valid() {
		return errors.New("occupant role must be a known OpenCaravan value")
	}
	if occupant.JoinedAt.IsZero() {
		return errors.New("joined_at must be set")
	}
	for i, clientAppID := range occupant.ClientAppIDs {
		if !clientAppID.Valid() {
			return fmt.Errorf("client_app_ids[%d] must be a valid UUID", i)
		}
	}
	return nil
}
