package opencaravan

import (
	"errors"
	"fmt"
	"time"
)

// Vehicle describes a physical vehicle that can carry one or more participants
// during a journey segment, together with the cert-backed metadata that gates
// who may drive it and who may edit the record.
//
// The record is owner-signed: OwnerUserID names the user whose enrolled client
// cert produced Integrity, and the signature covers the canonical encoding of
// every field except Integrity itself. Verifiers reproduce the canonical bytes
// via CanonicalEncoding and check the signature against the owner's enrolled
// cert.
//
// Edit authority is at the user level rather than the client_app level: any of
// the owner's enrolled client apps may produce a fresh Integrity over an
// updated record, so a user with multiple devices is not locked into the one
// that first uploaded.
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
	// OwnerUserID is the user whose enrolled client cert produced Integrity.
	// Permission to edit the record or update the authorized-drivers ACL is
	// scoped to this user, not to a specific client_app.
	OwnerUserID UUID `json:"owner_user_id"`
	// Capacity is the total number of possible occupants the vehicle can
	// carry, including the driver. A sedan is 5; a seven-seat minivan with
	// six seatbelts is 6.
	Capacity int `json:"capacity"`
	// AuthorizedDrivers names the users the owner has authorized to drive
	// this vehicle within the current journey. DriverAttestation values
	// validate against this list at the ACL version they recorded; see
	// VehicleACL for the version-evolution shape.
	AuthorizedDrivers []UUID `json:"authorized_drivers"`
	// ACLVersion is a monotonic counter incremented whenever
	// AuthorizedDrivers or EmergencyRule changes. Driver attestations
	// record the version they consulted so a later ACL revision does not
	// retroactively invalidate prior attestations.
	ACLVersion int `json:"acl_version"`
	// EmergencyRule is the owner-published fallback for when no one in
	// AuthorizedDrivers is available to drive. When set, a driver
	// attestation produced by a non-ACL participant is recorded with a
	// downgraded trust flag rather than rejected outright; when unset,
	// non-ACL attestations are recorded as ACL violations. Loss of trust
	// is information, never data loss.
	EmergencyRule *VehicleEmergencyRule `json:"emergency_rule,omitempty"`
	// Integrity is the owner's signature over CanonicalEncoding(). Optional
	// on a draft Vehicle that has not yet been signed; required on the
	// wire (the server rejects unsigned vehicle uploads).
	Integrity *Integrity `json:"integrity,omitempty"`
}

// VehicleEmergencyRule names the owner-published fallback semantics for when
// no AuthorizedDrivers participant is available to drive at a waypoint.
// The protocol records the rule so a future server-side or peer verifier can
// apply the same policy across implementations.
type VehicleEmergencyRule struct {
	// Kind names which fallback policy applies.
	Kind VehicleEmergencyRuleKind `json:"kind"`
}

// VehicleEmergencyRuleKind enumerates the fallback policies an owner may
// publish for emergency driver attestations.
type VehicleEmergencyRuleKind string

const (
	// VehicleEmergencyRuleNone means no emergency fallback is published.
	// A non-ACL driver attestation is recorded as an ACL violation.
	VehicleEmergencyRuleNone VehicleEmergencyRuleKind = "none"
	// VehicleEmergencyRuleAnyJourneyParticipant means any participant in
	// the journey may drive in an emergency. A non-ACL driver attestation
	// by a journey participant is recorded with a downgraded trust flag
	// rather than rejected.
	VehicleEmergencyRuleAnyJourneyParticipant VehicleEmergencyRuleKind = "any_journey_participant"
)

// Valid reports whether the rule kind is a known OpenCaravan value.
func (k VehicleEmergencyRuleKind) Valid() bool {
	switch k {
	case VehicleEmergencyRuleNone, VehicleEmergencyRuleAnyJourneyParticipant:
		return true
	default:
		return false
	}
}

// Validate reports whether the rule has a known kind.
func (r VehicleEmergencyRule) Validate() error {
	if !r.Kind.Valid() {
		return errors.New("emergency rule kind must be a known OpenCaravan value")
	}
	return nil
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

// Validate reports whether vehicle contains required identity, ownership,
// authorization, and signed-envelope shape, plus valid optional image
// resources. Structural only — signature cryptographic verification is the
// consumer's responsibility once the canonical bytes are reproduced via
// CanonicalEncoding.
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
	if !vehicle.OwnerUserID.Valid() {
		return errors.New("owner_user_id must be a valid UUID")
	}
	if vehicle.Capacity < 1 {
		return errors.New("capacity must be at least 1")
	}
	for i, driver := range vehicle.AuthorizedDrivers {
		if !driver.Valid() {
			return fmt.Errorf("authorized_drivers[%d] must be a valid UUID", i)
		}
	}
	if vehicle.ACLVersion < 1 {
		return errors.New("acl_version must be at least 1")
	}
	if vehicle.EmergencyRule != nil {
		if err := vehicle.EmergencyRule.Validate(); err != nil {
			return fmt.Errorf("emergency_rule: %w", err)
		}
	}
	if vehicle.Integrity != nil {
		if err := vehicle.Integrity.Validate(); err != nil {
			return fmt.Errorf("integrity: %w", err)
		}
	}
	return nil
}

// CanonicalEncoding returns the deterministic byte sequence the owner signs
// to produce Integrity. The Integrity field itself is excluded from the input
// (a signature cannot cover itself); every other field is included via
// [CanonicalJSON].
//
// Verifiers reproduce CanonicalEncoding on a received Vehicle, compute the
// signature input, and check Integrity.Signature against the cert identified
// by Integrity.KeyID. All conformant OpenCaravan implementations produce
// byte-identical output for the same input.
func (vehicle Vehicle) CanonicalEncoding() ([]byte, error) {
	cp := vehicle
	cp.Integrity = nil
	return CanonicalJSON(cp)
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
