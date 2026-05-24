package opencaravan

import (
	"errors"
	"fmt"
	"time"
)

// Vehicle describes a physical vehicle that can carry one or more participants
// during a journey segment.
type Vehicle struct {
	ID          UUID   `json:"id"`
	DisplayName string `json:"display_name"`
	Make        string `json:"make,omitempty"`
	Model       string `json:"model,omitempty"`
	ModelYear   int    `json:"model_year,omitempty"`
	Color       string `json:"color,omitempty"`
	// AvatarImage is the image clients can use for compact or map
	// representations of this vehicle.
	AvatarImage *ImageResourceRef `json:"avatar_image,omitempty"`
	// BannerImage is an optional wide image clients can use in richer vehicle
	// views.
	BannerImage *ImageResourceRef `json:"banner_image,omitempty"`
}

// SegmentVehicle describes one vehicle's participation in a journey segment.
type SegmentVehicle struct {
	ID        UUID              `json:"id"`
	SegmentID UUID              `json:"segment_id"`
	VehicleID UUID              `json:"vehicle_id"`
	Occupants []VehicleOccupant `json:"occupants"`
	Tracklog  []PositionSample  `json:"tracklog,omitempty"`
}

// OccupantRole describes what a journey participant is doing in a vehicle during
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

// VehicleOccupant links a journey participant to a vehicle during a journey
// segment, optionally naming the client apps acting on behalf of that
// participant. An occupant with no client_app_ids is a passenger who is not
// streaming telemetry from a personal device; an occupant with one or more
// client_app_ids may submit position samples through any of those apps.
type VehicleOccupant struct {
	JourneyParticipantID UUID         `json:"journey_participant_id"`
	ClientAppIDs         []UUID       `json:"client_app_ids,omitempty"`
	Role                 OccupantRole `json:"role"`
	JoinTime             time.Time    `json:"join_time"`
	LeaveTime            *time.Time   `json:"leave_time,omitempty"`
}

// Validate reports whether vehicle contains required identity and valid
// optional image resources.
func (vehicle Vehicle) Validate() error {
	if !vehicle.ID.Valid() {
		return errors.New("vehicle id must be a valid UUID")
	}
	if vehicle.DisplayName == "" {
		return errors.New("display_name must be set")
	}
	if vehicle.AvatarImage != nil {
		if err := vehicle.AvatarImage.Validate(); err != nil {
			return fmt.Errorf("avatar_image: %w", err)
		}
	}
	if vehicle.BannerImage != nil {
		if err := vehicle.BannerImage.Validate(); err != nil {
			return fmt.Errorf("banner_image: %w", err)
		}
	}
	return nil
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
		participants[occupant.JourneyParticipantID] = struct{}{}
	}
	for i, sample := range vehicle.Tracklog {
		if err := sample.Validate(); err != nil {
			return fmt.Errorf("tracklog sample %d: %w", i, err)
		}
		if sample.SegmentVehicleID != vehicle.ID {
			return fmt.Errorf("tracklog sample %d: segment_vehicle_id does not match vehicle", i)
		}
		if sample.SegmentID != vehicle.SegmentID {
			return fmt.Errorf("tracklog sample %d: segment_id does not match segment_vehicle", i)
		}
		if _, ok := participants[sample.JourneyParticipantID]; !ok {
			return fmt.Errorf("tracklog sample %d: journey_participant_id is not an occupant", i)
		}
	}

	return nil
}

// Validate reports whether occupant contains a participant, valid role, and
// join time.
func (occupant VehicleOccupant) Validate() error {
	if !occupant.JourneyParticipantID.Valid() {
		return errors.New("journey participant id must be a valid UUID")
	}
	if !occupant.Role.Valid() {
		return errors.New("occupant role must be a known OpenCaravan value")
	}
	if occupant.JoinTime.IsZero() {
		return errors.New("join_time must be set")
	}
	for i, clientAppID := range occupant.ClientAppIDs {
		if !clientAppID.Valid() {
			return fmt.Errorf("client_app_ids[%d] must be a valid UUID", i)
		}
	}
	return nil
}
