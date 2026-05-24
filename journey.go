package opencaravan

import (
	"errors"
	"fmt"
	"time"
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

// Valid reports whether the journey state is a known OpenCaravan value.
func (s JourneyState) Valid() bool {
	switch s {
	case JourneyPlanned, JourneyActive, JourneyClosed, JourneyExpired, JourneyDeleted:
		return true
	default:
		return false
	}
}

// Journey describes a private, invite-only group drive shared through
// OpenCaravan.
//
// Journey is the aggregate representation. APIs may return participants,
// segments, and media inline for small journeys, or page those collections
// separately for large retained journeys. DeletionTime is an immutable scheduled
// hard-deletion time; nil means no scheduled deletion. New participants join
// through server-issued invite tokens, and invite creation is governed by
// each attached participant's privileges.
type Journey struct {
	ID              UUID   `json:"id"`
	OriginServerURL string `json:"origin_server_url"`
	Title           string `json:"title"`
	Description     string `json:"description,omitempty"`
	// AvatarImage is the image clients can use for compact journey
	// representations, including invite previews.
	AvatarImage *ImageResourceRef `json:"avatar_image,omitempty"`
	// BannerImage is an optional wide image clients can use in richer journey
	// views and invite previews.
	BannerImage      *ImageResourceRef    `json:"banner_image,omitempty"`
	State            JourneyState         `json:"state"`
	DeletionTime     *time.Time           `json:"deletion_time,omitempty"`
	Features         JourneyFeatures      `json:"features"`
	Participants     []JourneyParticipant `json:"participants,omitempty"`
	Segments         []JourneySegment     `json:"segments,omitempty"`
	SharedMedia      []SharedMedia        `json:"shared_media,omitempty"`
	CreationTime     time.Time            `json:"creation_time"`
	PlannedStartTime *time.Time           `json:"planned_start_time,omitempty"`
	ActualStartTime  *time.Time           `json:"actual_start_time,omitempty"`
	TrackingEndTime  *time.Time           `json:"tracking_end_time,omitempty"`
}

// JourneyFeatures describes optional capabilities enabled for a journey.
type JourneyFeatures struct {
	ExportAllowed bool `json:"export_allowed"`
	MediaAllowed  bool `json:"media_allowed"`
}

// JourneyParticipant describes a user's membership in one journey.
//
// Segment vehicle occupants describe who is in a vehicle for a bounded segment.
// JourneyParticipant describes the journey-level membership, the optional
// journey-visible profile projection, and the privileges that membership
// carries.
type JourneyParticipant struct {
	ID         UUID                         `json:"id"`
	JourneyID  UUID                         `json:"journey_id"`
	UserID     UUID                         `json:"user_id"`
	Profile    *UserProfile                 `json:"profile,omitempty"`
	Privileges JourneyParticipantPrivileges `json:"privileges"`
	JoinTime   time.Time                    `json:"join_time"`
	LeaveTime  *time.Time                   `json:"leave_time,omitempty"`
}

// JourneyParticipantPrivileges describes what a participant may do within a
// journey.
type JourneyParticipantPrivileges struct {
	InviteGeneration *InviteGenerationPermissions `json:"invite_generation,omitempty"`
}

// ClientApp describes one OpenCaravan-capable app installation or session
// acting on behalf of a user.
type ClientApp struct {
	ID           UUID       `json:"id"`
	UserID       UUID       `json:"user_id"`
	Name         string     `json:"name"`
	Version      string     `json:"version,omitempty"`
	Platform     string     `json:"platform,omitempty"`
	DeviceName   string     `json:"device_name,omitempty"`
	Capabilities []string   `json:"capabilities,omitempty"`
	LastSeenTime *time.Time `json:"last_seen_time,omitempty"`
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

// Valid reports whether the segment state is a known OpenCaravan value.
func (s SegmentState) Valid() bool {
	switch s {
	case SegmentPlanned, SegmentActive, SegmentClosed, SegmentDiscarded:
		return true
	default:
		return false
	}
}

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
	StartTime time.Time        `json:"start_time"`
	EndTime   *time.Time       `json:"end_time,omitempty"`
}

// Validate reports whether journey contains the required identity, immutable
// deletion timestamp, timestamps, and aggregate relationships.
func (journey Journey) Validate() error {
	if !journey.ID.Valid() {
		return errors.New("journey id must be a valid UUID")
	}
	if journey.OriginServerURL == "" {
		return errors.New("origin_server_url must be set")
	}
	if journey.Title == "" {
		return errors.New("title must be set")
	}
	if !journey.State.Valid() {
		return errors.New("state must be a known OpenCaravan value")
	}
	if journey.AvatarImage != nil {
		if err := journey.AvatarImage.Validate(); err != nil {
			return fmt.Errorf("avatar_image: %w", err)
		}
	}
	if journey.BannerImage != nil {
		if err := journey.BannerImage.Validate(); err != nil {
			return fmt.Errorf("banner_image: %w", err)
		}
	}
	if journey.DeletionTime != nil && journey.DeletionTime.IsZero() {
		return errors.New("deletion_time must be a non-zero time")
	}
	if journey.CreationTime.IsZero() {
		return errors.New("creation_time must be set")
	}
	if journey.DeletionTime != nil && journey.DeletionTime.Before(journey.CreationTime) {
		return errors.New("deletion_time must not be before creation_time")
	}
	if journey.PlannedStartTime != nil && journey.PlannedStartTime.IsZero() {
		return errors.New("planned_start_time must be a non-zero time")
	}
	if journey.ActualStartTime != nil && journey.ActualStartTime.IsZero() {
		return errors.New("actual_start_time must be a non-zero time")
	}
	if journey.TrackingEndTime != nil && journey.TrackingEndTime.IsZero() {
		return errors.New("tracking_end_time must be a non-zero time")
	}

	for i, participant := range journey.Participants {
		if err := participant.Validate(); err != nil {
			return fmt.Errorf("participants[%d]: %w", i, err)
		}
		if participant.JourneyID != journey.ID {
			return fmt.Errorf("participants[%d]: journey_id does not match journey", i)
		}
	}
	for i, segment := range journey.Segments {
		if err := segment.Validate(); err != nil {
			return fmt.Errorf("segments[%d]: %w", i, err)
		}
		if segment.JourneyID != journey.ID {
			return fmt.Errorf("segments[%d]: journey_id does not match journey", i)
		}
	}
	for i, media := range journey.SharedMedia {
		if !media.ID.Valid() {
			return fmt.Errorf("shared_media[%d]: id must be a valid UUID", i)
		}
		if media.JourneyID != journey.ID {
			return fmt.Errorf("shared_media[%d]: journey_id does not match journey", i)
		}
	}

	return nil
}

// Validate reports whether segment has the required identity, state,
// timestamps, and nested segment vehicles. Each nested SegmentVehicle must
// also reference this segment's ID so a vehicle from another segment cannot
// silently attach.
func (segment JourneySegment) Validate() error {
	if !segment.ID.Valid() {
		return errors.New("id must be a valid UUID")
	}
	if !segment.JourneyID.Valid() {
		return errors.New("journey_id must be a valid UUID")
	}
	if !segment.State.Valid() {
		return errors.New("state must be a known OpenCaravan value")
	}
	if segment.StartTime.IsZero() {
		return errors.New("start_time must be set")
	}
	if segment.EndTime != nil {
		if segment.EndTime.IsZero() {
			return errors.New("end_time must be a non-zero time")
		}
		if segment.EndTime.Before(segment.StartTime) {
			return errors.New("end_time must not be before start_time")
		}
	}
	for i, sv := range segment.Vehicles {
		if err := sv.Validate(); err != nil {
			return fmt.Errorf("vehicles[%d]: %w", i, err)
		}
		if sv.SegmentID != segment.ID {
			return fmt.Errorf("vehicles[%d]: segment_id does not match segment", i)
		}
	}
	return nil
}

// Validate reports whether participant has the required journey membership,
// server-scoped user ID, optional profile projection, and join time.
func (participant JourneyParticipant) Validate() error {
	if !participant.ID.Valid() {
		return errors.New("journey participant id must be a valid UUID")
	}
	if !participant.JourneyID.Valid() {
		return errors.New("journey_id must be a valid UUID")
	}
	if !participant.UserID.Valid() {
		return errors.New("user_id must be a valid UUID")
	}
	if participant.Profile != nil {
		if err := participant.Profile.Validate(); err != nil {
			return fmt.Errorf("profile: %w", err)
		}
	}
	if err := participant.Privileges.Validate(); err != nil {
		return fmt.Errorf("privileges: %w", err)
	}
	if participant.JoinTime.IsZero() {
		return errors.New("join_time must be set")
	}
	return nil
}

// Validate reports whether privileges contain valid optional capability
// envelopes.
func (privileges JourneyParticipantPrivileges) Validate() error {
	if privileges.InviteGeneration != nil {
		if err := privileges.InviteGeneration.Validate(); err != nil {
			return fmt.Errorf("invite_generation: %w", err)
		}
	}
	return nil
}

// Validate reports whether app has the required app, user, and display fields.
func (app ClientApp) Validate() error {
	if !app.ID.Valid() {
		return errors.New("client app id must be a valid UUID")
	}
	if !app.UserID.Valid() {
		return errors.New("user_id must be a valid UUID")
	}
	if app.Name == "" {
		return errors.New("name must be set")
	}
	return nil
}
