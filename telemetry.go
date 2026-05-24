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
	JourneyID              string    `json:"journey_id"`
	ParticipantID          string    `json:"participant_id"`
	DeviceID               string    `json:"device_id,omitempty"`
	ClientSequence         int64     `json:"client_sequence"`
	CapturedAt             time.Time `json:"captured_at"`
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

// Validate reports whether the sample contains a valid sequence, capture time,
// and latitude/longitude pair.
func (s PositionSample) Validate() error {
	if s.ClientSequence < 0 {
		return errors.New("client sequence must be non-negative")
	}
	if s.CapturedAt.IsZero() {
		return errors.New("captured_at must be set")
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
