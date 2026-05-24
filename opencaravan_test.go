package opencaravan

import (
	"encoding/json"
	"testing"
	"time"
)

func TestRegistrationModeValid(t *testing.T) {
	tests := []struct {
		name string
		mode RegistrationMode
		want bool
	}{
		{name: "closed", mode: RegistrationClosed, want: true},
		{name: "invite", mode: RegistrationInvite, want: true},
		{name: "open", mode: RegistrationOpen, want: true},
		{name: "unknown", mode: RegistrationMode("unknown"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mode.Valid(); got != tt.want {
				t.Fatalf("Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJourneyInvitePayloadJSON(t *testing.T) {
	expiresAt := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	payload := NewJourneyInvitePayload("https://public.spivot.net", "journey_123", "token", expiresAt)
	payload.PolicyHash = "sha256:abc"
	payload.DisplayName = "Sunday Ridge Drive"

	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded JourneyInvitePayload
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if decoded.Type != JourneyInvitePayloadType {
		t.Fatalf("Type = %q, want %q", decoded.Type, JourneyInvitePayloadType)
	}
	if decoded.Version != JourneyInvitePayloadVersion {
		t.Fatalf("Version = %d, want %d", decoded.Version, JourneyInvitePayloadVersion)
	}
	if !decoded.ExpiresAt.Equal(expiresAt) {
		t.Fatalf("ExpiresAt = %s, want %s", decoded.ExpiresAt, expiresAt)
	}
}

func TestPositionSampleValidate(t *testing.T) {
	valid := PositionSample{
		JourneyID:        "journey_123",
		ParticipantID:    "participant_123",
		ClientSequence:   1,
		CapturedAt:       time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC),
		LatitudeE7:       451234567,
		LongitudeE7:      -1221234567,
		HeadingDegreesE2: ptr[int32](35999),
	}

	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*PositionSample)
	}{
		{name: "negative sequence", mutate: func(s *PositionSample) { s.ClientSequence = -1 }},
		{name: "missing captured time", mutate: func(s *PositionSample) { s.CapturedAt = time.Time{} }},
		{name: "latitude too low", mutate: func(s *PositionSample) { s.LatitudeE7 = -900000001 }},
		{name: "latitude too high", mutate: func(s *PositionSample) { s.LatitudeE7 = 900000001 }},
		{name: "longitude too low", mutate: func(s *PositionSample) { s.LongitudeE7 = -1800000001 }},
		{name: "longitude too high", mutate: func(s *PositionSample) { s.LongitudeE7 = 1800000001 }},
		{name: "heading too high", mutate: func(s *PositionSample) { s.HeadingDegreesE2 = ptr[int32](36000) }},
		{name: "battery too high", mutate: func(s *PositionSample) { s.BatteryLevelPermille = ptr[int32](1001) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sample := valid
			tt.mutate(&sample)
			if err := sample.Validate(); err == nil {
				t.Fatal("Validate() error = nil, want error")
			}
		})
	}
}

func ptr[T any](value T) *T {
	return &value
}
