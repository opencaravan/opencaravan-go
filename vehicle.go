package opencaravan

import (
	"errors"
	"fmt"
	"time"
)

// Vehicle is the journey-scoped, owner-signed metadata bundle for
// a physical vehicle taking part in a trip: who owns the record,
// what the vehicle looks like, what photos to render, and the
// monotonic revision counter the owner increments whenever any
// of that metadata changes.
//
// Vehicle is intentionally opaque to the server. The protocol
// keeps the vehicle's *descriptive* state — make/model/photos/
// notes — in this signed bundle so any client that holds the
// bundle has the full picture, while the *authorization* state
// (who may drive in this journey, what the emergency fallback
// is) lives on the separate [VehicleACL] type. That split lets
// clients gossip a metadata change (a new photo, a corrected
// model year) independently of an authorization change (adding a
// driver), which matters in the offline-first model where each
// kind of change has different timing and verification needs.
//
// The journey-scoped Vehicle is constructed by copying display
// name, photos, make/model, and capacity from a [GarageVehicle]
// the user has selected from their persistent garage, then
// publishing it fresh for this journey. There is no wire-level
// link from Vehicle back to its GarageVehicle source.
//
// Each Vehicle value is one revision in a monotonic chain.
// RevisionVersion starts at 1 and is strictly increasing per ID.
// Revisions are signed by the OwnerUserID's enrolled client cert
// (any of the owner's enrolled apps may sign; edit authority is
// user-scoped, not client-app-scoped). The server retains the
// full revision history so a recipient can audit how a vehicle's
// metadata evolved during the journey.
type Vehicle struct {
	ID UUID `json:"id"`
	// OwnerUserID is the user whose enrolled client cert produced
	// Integrity. Permission to publish a new revision is scoped to
	// this user, not to a specific client_app.
	OwnerUserID UUID `json:"owner_user_id"`
	// RevisionVersion is the monotonic counter for this vehicle's
	// metadata bundle. Starts at 1, strictly increases per ID.
	// Independent of [VehicleACL.ACLVersion] — metadata and
	// authorization version separately.
	RevisionVersion int `json:"revision_version"`
	// RevisionTime is when the owner signed this revision. Used by
	// recipients to render "last updated" and to break ties when
	// two revisions arrive concurrently.
	RevisionTime time.Time `json:"revision_time"`
	DisplayName  string    `json:"display_name"`
	Make         string    `json:"make,omitempty"`
	Model        string    `json:"model,omitempty"`
	ModelYear    int       `json:"model_year,omitempty"`
	Color        string    `json:"color,omitempty"`
	// Capacity is the total number of possible occupants the
	// vehicle can carry, including the driver. A sedan is 5; a
	// seven-seat minivan with six seatbelts is 6.
	Capacity int `json:"capacity"`
	// AvatarBlob references the compact / map-tile photo for this
	// vehicle, content-addressed via [BlobRef]. The bytes are
	// uploaded to the server's blob layer; multiple Vehicle
	// revisions referencing the same photo deduplicate by hash.
	AvatarBlob *BlobRef `json:"avatar_blob,omitempty"`
	// BannerBlob references the wide / detail-view photo, same
	// content-addressed model as AvatarBlob.
	BannerBlob *BlobRef `json:"banner_blob,omitempty"`
	// Notes is owner-authored free text. Surface in client UIs as
	// "owner notes" or similar.
	Notes string `json:"notes,omitempty"`
	// Integrity is the owner's signature over CanonicalEncoding().
	// Optional on a draft Vehicle that has not yet been signed;
	// required on the wire — the server rejects unsigned uploads.
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

// Validate reports whether vehicle contains required identity,
// ownership, revision-bearing fields, capacity, and signed-envelope
// shape, plus valid optional blob references. Structural only —
// cryptographic verification of Integrity is the consumer's
// responsibility once canonical bytes are reproduced via
// CanonicalEncoding.
func (vehicle Vehicle) Validate() error {
	if !vehicle.ID.Valid() {
		return errors.New("vehicle id must be a valid UUID")
	}
	if !vehicle.OwnerUserID.Valid() {
		return errors.New("owner_user_id must be a valid UUID")
	}
	if vehicle.RevisionVersion < 1 {
		return errors.New("revision_version must be at least 1")
	}
	if vehicle.RevisionTime.IsZero() {
		return errors.New("revision_time must be set")
	}
	if vehicle.DisplayName == "" {
		return errors.New("display_name must be set")
	}
	if vehicle.Capacity < 1 {
		return errors.New("capacity must be at least 1")
	}
	if vehicle.AvatarBlob != nil {
		if err := vehicle.AvatarBlob.Validate(); err != nil {
			return fmt.Errorf("avatar_blob: %w", err)
		}
	}
	if vehicle.BannerBlob != nil {
		if err := vehicle.BannerBlob.Validate(); err != nil {
			return fmt.Errorf("banner_blob: %w", err)
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
