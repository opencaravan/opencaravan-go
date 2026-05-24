package opencaravan

import (
	"encoding/json"
	"testing"
	"time"
)

const (
	testJourneyID        UUID = "11111111-1111-4111-8111-111111111111"
	testSegmentID        UUID = "22222222-2222-4222-8222-222222222222"
	testSegmentVehicleID UUID = "33333333-3333-4333-8333-333333333333"
	testVehicleID        UUID = "44444444-4444-4444-8444-444444444444"
	testParticipantID    UUID = "55555555-5555-4555-8555-555555555555"
	testClientAppID      UUID = "66666666-6666-4666-8666-666666666666"
)

func TestUUIDMarshalTextRequiresCanonicalID(t *testing.T) {
	tests := []struct {
		name    string
		id      UUID
		wantErr bool
	}{
		{name: "valid", id: testJourneyID},
		{name: "uppercase valid", id: UUID("11111111-1111-4111-8111-AAAAAAAAAAAA")},
		{name: "missing hyphens", id: UUID("11111111111141118111111111111111"), wantErr: true},
		{name: "nil uuid", id: UUID(nilUUID), wantErr: true},
		{name: "empty", id: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.id.MarshalText()
			if (err != nil) != tt.wantErr {
				t.Fatalf("MarshalText() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUUIDJSONIsStrictAndCanonical(t *testing.T) {
	encoded, err := json.Marshal(UUID("11111111-1111-4111-8111-AAAAAAAAAAAA"))
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if got, want := string(encoded), `"11111111-1111-4111-8111-aaaaaaaaaaaa"`; got != want {
		t.Fatalf("Marshal() = %s, want %s", got, want)
	}

	var id UUID
	if err := json.Unmarshal([]byte(`"not-a-uuid"`), &id); err == nil {
		t.Fatal("Unmarshal() error = nil, want error")
	}
}

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
	payload := NewJourneyInvitePayload("https://public.spivot.net", testJourneyID, "token", expiresAt)
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
	if decoded.JourneyID != testJourneyID {
		t.Fatalf("JourneyID = %q, want %q", decoded.JourneyID, testJourneyID)
	}
	if !decoded.ExpiresAt.Equal(expiresAt) {
		t.Fatalf("ExpiresAt = %s, want %s", decoded.ExpiresAt, expiresAt)
	}
}

func TestPositionSampleValidate(t *testing.T) {
	valid := validPositionSample()

	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*PositionSample)
	}{
		{name: "missing journey id", mutate: func(s *PositionSample) { s.JourneyID = "" }},
		{name: "missing segment id", mutate: func(s *PositionSample) { s.SegmentID = "" }},
		{name: "missing segment vehicle id", mutate: func(s *PositionSample) { s.SegmentVehicleID = "" }},
		{name: "missing participant id", mutate: func(s *PositionSample) { s.ParticipantID = "" }},
		{name: "missing client app id", mutate: func(s *PositionSample) { s.ClientAppID = "" }},
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

func TestSegmentVehicleValidate(t *testing.T) {
	valid := SegmentVehicle{
		ID:        testSegmentVehicleID,
		SegmentID: testSegmentID,
		VehicleID: testVehicleID,
		Occupants: []VehicleOccupant{
			{
				ParticipantID: testParticipantID,
				ClientAppIDs:  []UUID{testClientAppID},
				Role:          OccupantDriver,
				JoinedAt:      time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC),
			},
		},
		Tracklog: []PositionSample{validPositionSample()},
	}

	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*SegmentVehicle)
	}{
		{name: "missing segment vehicle id", mutate: func(v *SegmentVehicle) { v.ID = "" }},
		{name: "missing segment id", mutate: func(v *SegmentVehicle) { v.SegmentID = "" }},
		{name: "missing vehicle id", mutate: func(v *SegmentVehicle) { v.VehicleID = "" }},
		{name: "no occupants", mutate: func(v *SegmentVehicle) { v.Occupants = nil }},
		{name: "unknown occupant role", mutate: func(v *SegmentVehicle) { v.Occupants[0].Role = OccupantRole("pilot") }},
		{name: "tracklog from non occupant", mutate: func(v *SegmentVehicle) { v.Tracklog[0].ParticipantID = testVehicleID }},
		{name: "tracklog for other segment vehicle", mutate: func(v *SegmentVehicle) { v.Tracklog[0].SegmentVehicleID = testVehicleID }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vehicle := valid
			vehicle.Occupants = append([]VehicleOccupant(nil), valid.Occupants...)
			vehicle.Tracklog = append([]PositionSample(nil), valid.Tracklog...)
			tt.mutate(&vehicle)
			if err := vehicle.Validate(); err == nil {
				t.Fatal("Validate() error = nil, want error")
			}
		})
	}
}

func validPositionSample() PositionSample {
	return PositionSample{
		JourneyID:        testJourneyID,
		SegmentID:        testSegmentID,
		SegmentVehicleID: testSegmentVehicleID,
		ParticipantID:    testParticipantID,
		ClientAppID:      testClientAppID,
		ClientSequence:   1,
		CapturedAt:       time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC),
		LatitudeE7:       451234567,
		LongitudeE7:      -1221234567,
		HeadingDegreesE2: ptr[int32](35999),
	}
}

func ptr[T any](value T) *T {
	return &value
}
