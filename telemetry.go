package opencaravan

import (
	"errors"
	"time"
)

const (
	minLatitudeE7  = -900000000
	maxLatitudeE7  = 900000000
	minLongitudeE7 = -1800000000
	maxLongitudeE7 = 1800000000
)

// PositionSample is one captured participant position update.
type PositionSample struct {
	JourneyID              UUID      `json:"journey_id"`
	SegmentID              UUID      `json:"segment_id"`
	SegmentVehicleID       UUID      `json:"segment_vehicle_id"`
	JourneyParticipantID   UUID      `json:"journey_participant_id"`
	ClientAppID            UUID      `json:"client_app_id"`
	ClientSequence         int64     `json:"client_sequence"`
	CaptureTime            time.Time `json:"capture_time"`
	LatitudeE7             int32     `json:"latitude_e7"`
	LongitudeE7            int32     `json:"longitude_e7"`
	AltitudeMM             *int64    `json:"altitude_mm,omitempty"`
	HorizontalAccuracyMM   *int64    `json:"horizontal_accuracy_mm,omitempty"`
	VerticalAccuracyMM     *int64    `json:"vertical_accuracy_mm,omitempty"`
	SpeedMMS               *int64    `json:"speed_mm_s,omitempty"`
	HeadingDegreesE2       *int32    `json:"heading_deg_e2,omitempty"`
	BatteryLevelPermille   *int32    `json:"battery_level_permille,omitempty"`
	Source                 string    `json:"source,omitempty"`
	MotionState            string    `json:"motion_state,omitempty"`
	ClientMetadataRevision string    `json:"client_metadata_revision,omitempty"`
}

// Validate reports whether the sample contains valid object identities,
// sequence data, capture time, and latitude/longitude pair.
func (s PositionSample) Validate() error {
	if !s.JourneyID.Valid() {
		return errors.New("journey_id must be a valid UUID")
	}
	if !s.SegmentID.Valid() {
		return errors.New("segment_id must be a valid UUID")
	}
	if !s.SegmentVehicleID.Valid() {
		return errors.New("segment_vehicle_id must be a valid UUID")
	}
	if !s.JourneyParticipantID.Valid() {
		return errors.New("journey_participant_id must be a valid UUID")
	}
	if !s.ClientAppID.Valid() {
		return errors.New("client_app_id must be a valid UUID")
	}
	if s.ClientSequence < 0 {
		return errors.New("client sequence must be non-negative")
	}
	if s.CaptureTime.IsZero() {
		return errors.New("capture_time must be set")
	}
	if s.LatitudeE7 < minLatitudeE7 || s.LatitudeE7 > maxLatitudeE7 {
		return errors.New("latitude_e7 out of range")
	}
	if s.LongitudeE7 < minLongitudeE7 || s.LongitudeE7 > maxLongitudeE7 {
		return errors.New("longitude_e7 out of range")
	}
	if s.HeadingDegreesE2 != nil && (*s.HeadingDegreesE2 < 0 || *s.HeadingDegreesE2 >= 36000) {
		return errors.New("heading_deg_e2 out of range")
	}
	if s.BatteryLevelPermille != nil && (*s.BatteryLevelPermille < 0 || *s.BatteryLevelPermille > 1000) {
		return errors.New("battery_level_permille out of range")
	}
	return nil
}
