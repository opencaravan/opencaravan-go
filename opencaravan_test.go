package opencaravan

import (
	"encoding/base64"
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
	testMembershipID     UUID = "77777777-7777-4777-8777-777777777777"
	testInviteID         UUID = "88888888-8888-4888-8888-888888888888"
	testMediaID          UUID = "99999999-9999-4999-8999-999999999999"
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

func TestJourneyRetentionModeValid(t *testing.T) {
	tests := []struct {
		name string
		mode JourneyRetentionMode
		want bool
	}{
		{name: "ephemeral", mode: JourneyRetentionEphemeral, want: true},
		{name: "download window", mode: JourneyRetentionDownloadWindow, want: true},
		{name: "forever", mode: JourneyRetentionForever, want: true},
		{name: "unknown", mode: JourneyRetentionMode("unknown"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mode.Valid(); got != tt.want {
				t.Fatalf("Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServerPolicyValidate(t *testing.T) {
	valid := validServerPolicy()

	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*ServerPolicy)
	}{
		{name: "missing protocol version", mutate: func(p *ServerPolicy) { p.ProtocolVersion = "" }},
		{name: "missing server url", mutate: func(p *ServerPolicy) { p.ServerURL = "" }},
		{name: "unknown registration mode", mutate: func(p *ServerPolicy) { p.RegistrationMode = RegistrationMode("unknown") }},
		{name: "no journey policies", mutate: func(p *ServerPolicy) { p.JourneyPolicies = nil }},
		{name: "duplicate policy id", mutate: func(p *ServerPolicy) {
			p.JourneyPolicies[1].ID = p.JourneyPolicies[0].ID
		}},
		{name: "missing default policy", mutate: func(p *ServerPolicy) { p.DefaultJourneyPolicyID = "missing" }},
		{name: "ephemeral profile with download window", mutate: func(p *ServerPolicy) {
			p.JourneyPolicies[0].MaxDownloadWindow = "24h"
		}},
		{name: "ephemeral profile with export", mutate: func(p *ServerPolicy) {
			p.JourneyPolicies[0].ExportSupported = true
		}},
		{name: "download-window profile without export", mutate: func(p *ServerPolicy) {
			p.JourneyPolicies[1].ExportSupported = false
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := valid
			policy.JourneyPolicies = append([]JourneyPolicyProfile(nil), valid.JourneyPolicies...)
			tt.mutate(&policy)
			if err := policy.Validate(); err == nil {
				t.Fatal("Validate() error = nil, want error")
			}
		})
	}
}

func TestJourneyPolicyValidate(t *testing.T) {
	activeExpiresAt := time.Date(2026, 5, 24, 18, 0, 0, 0, time.UTC)
	downloadUntil := time.Date(2026, 5, 31, 18, 0, 0, 0, time.UTC)
	purgeAt := time.Date(2026, 6, 1, 18, 0, 0, 0, time.UTC)

	valid := JourneyPolicy{
		PolicyID:        "public-download-7d",
		PolicyHash:      "sha256:abc",
		RetentionMode:   JourneyRetentionDownloadWindow,
		ActiveExpiresAt: &activeExpiresAt,
		DownloadUntil:   &downloadUntil,
		PurgeAt:         &purgeAt,
		ExportSupported: true,
		MediaAllowed:    true,
	}

	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	ephemeral := JourneyPolicy{
		PolicyID:      "public-ephemeral",
		PolicyHash:    "sha256:def",
		RetentionMode: JourneyRetentionEphemeral,
		PurgeAt:       &purgeAt,
	}
	if err := ephemeral.Validate(); err != nil {
		t.Fatalf("ephemeral Validate() error = %v", err)
	}

	forever := JourneyPolicy{
		PolicyID:        "private-forever",
		PolicyHash:      "sha256:ghi",
		RetentionMode:   JourneyRetentionForever,
		RetainForever:   true,
		ExportSupported: true,
	}
	if err := forever.Validate(); err != nil {
		t.Fatalf("forever Validate() error = %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*JourneyPolicy)
	}{
		{name: "missing policy id", mutate: func(p *JourneyPolicy) { p.PolicyID = "" }},
		{name: "missing policy hash", mutate: func(p *JourneyPolicy) { p.PolicyHash = "" }},
		{name: "zero active expiration", mutate: func(p *JourneyPolicy) { p.ActiveExpiresAt = ptr(time.Time{}) }},
		{name: "purge before download", mutate: func(p *JourneyPolicy) {
			beforeDownload := downloadUntil.Add(-time.Hour)
			p.PurgeAt = &beforeDownload
		}},
		{name: "download window without download until", mutate: func(p *JourneyPolicy) { p.DownloadUntil = nil }},
		{name: "download window without purge", mutate: func(p *JourneyPolicy) { p.PurgeAt = nil }},
		{name: "download window without export", mutate: func(p *JourneyPolicy) { p.ExportSupported = false }},
		{name: "ephemeral without purge", mutate: func(p *JourneyPolicy) {
			p.RetentionMode = JourneyRetentionEphemeral
			p.DownloadUntil = nil
			p.PurgeAt = nil
		}},
		{name: "ephemeral retaining forever", mutate: func(p *JourneyPolicy) {
			p.RetentionMode = JourneyRetentionEphemeral
			p.DownloadUntil = nil
			p.RetainForever = true
		}},
		{name: "ephemeral with download until", mutate: func(p *JourneyPolicy) {
			p.RetentionMode = JourneyRetentionEphemeral
		}},
		{name: "ephemeral with export", mutate: func(p *JourneyPolicy) {
			p.RetentionMode = JourneyRetentionEphemeral
			p.DownloadUntil = nil
			p.ExportSupported = true
		}},
		{name: "forever without retain flag", mutate: func(p *JourneyPolicy) {
			p.RetentionMode = JourneyRetentionForever
			p.DownloadUntil = nil
			p.PurgeAt = nil
			p.RetainForever = false
		}},
		{name: "forever with purge", mutate: func(p *JourneyPolicy) {
			p.RetentionMode = JourneyRetentionForever
			p.DownloadUntil = nil
			p.RetainForever = true
		}},
		{name: "forever without export", mutate: func(p *JourneyPolicy) {
			p.RetentionMode = JourneyRetentionForever
			p.DownloadUntil = nil
			p.PurgeAt = nil
			p.RetainForever = true
			p.ExportSupported = false
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := valid
			tt.mutate(&policy)
			if err := policy.Validate(); err == nil {
				t.Fatal("Validate() error = nil, want error")
			}
		})
	}
}

func TestJourneyStateValid(t *testing.T) {
	tests := []struct {
		name  string
		state JourneyState
		want  bool
	}{
		{name: "planned", state: JourneyPlanned, want: true},
		{name: "active", state: JourneyActive, want: true},
		{name: "closed", state: JourneyClosed, want: true},
		{name: "expired", state: JourneyExpired, want: true},
		{name: "deleted", state: JourneyDeleted, want: true},
		{name: "unknown", state: JourneyState("unknown"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.Valid(); got != tt.want {
				t.Fatalf("Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJourneyValidate(t *testing.T) {
	valid := validJourney()

	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*Journey)
	}{
		{name: "missing journey id", mutate: func(j *Journey) { j.ID = "" }},
		{name: "missing origin server", mutate: func(j *Journey) { j.OriginServerURL = "" }},
		{name: "missing title", mutate: func(j *Journey) { j.Title = "" }},
		{name: "unknown state", mutate: func(j *Journey) { j.State = JourneyState("unknown") }},
		{name: "invalid policy", mutate: func(j *Journey) { j.Policy.PolicyHash = "" }},
		{name: "missing created at", mutate: func(j *Journey) { j.CreatedAt = time.Time{} }},
		{name: "zero starts at", mutate: func(j *Journey) { j.StartsAt = ptr(time.Time{}) }},
		{name: "participant for other journey", mutate: func(j *Journey) { j.Participants[0].JourneyID = testVehicleID }},
		{name: "segment for other journey", mutate: func(j *Journey) { j.Segments[0].JourneyID = testVehicleID }},
		{name: "media for other journey", mutate: func(j *Journey) { j.SharedMedia[0].JourneyID = testVehicleID }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			journey := valid
			journey.Participants = append([]JourneyParticipant(nil), valid.Participants...)
			journey.Segments = append([]JourneySegment(nil), valid.Segments...)
			journey.SharedMedia = append([]SharedMedia(nil), valid.SharedMedia...)
			tt.mutate(&journey)
			if err := journey.Validate(); err == nil {
				t.Fatal("Validate() error = nil, want error")
			}
		})
	}
}

func TestNewJourneyInviteToken(t *testing.T) {
	expiresAt := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)

	singleUse, err := NewJourneyInviteToken(JourneyInviteSingleUse, expiresAt)
	if err != nil {
		t.Fatalf("NewJourneyInviteToken() error = %v", err)
	}
	if singleUse.UseMode != JourneyInviteSingleUse {
		t.Fatalf("UseMode = %q, want %q", singleUse.UseMode, JourneyInviteSingleUse)
	}
	if singleUse.MaxUses != 1 {
		t.Fatalf("MaxUses = %d, want 1", singleUse.MaxUses)
	}
	tokenBytes, err := base64.RawURLEncoding.DecodeString(singleUse.Value)
	if err != nil {
		t.Fatalf("DecodeString() error = %v", err)
	}
	if len(tokenBytes) != journeyInviteTokenBytes {
		t.Fatalf("decoded token length = %d, want %d", len(tokenBytes), journeyInviteTokenBytes)
	}

	multiUse, err := NewJourneyInviteToken(JourneyInviteMultiUse, expiresAt)
	if err != nil {
		t.Fatalf("NewJourneyInviteToken() error = %v", err)
	}
	if multiUse.MaxUses != 0 {
		t.Fatalf("MaxUses = %d, want 0 for uncapped multi-use", multiUse.MaxUses)
	}

	if _, err := NewJourneyInviteToken(JourneyInviteUseMode("unknown"), expiresAt); err == nil {
		t.Fatal("NewJourneyInviteToken() error = nil, want invalid use mode error")
	}
	if _, err := NewJourneyInviteToken(JourneyInviteSingleUse, time.Time{}); err == nil {
		t.Fatal("NewJourneyInviteToken() error = nil, want missing expires_at error")
	}
}

func TestJourneyInviteJSONAndValidate(t *testing.T) {
	invite := validJourneyInvite(t)

	if err := invite.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	encoded, err := json.Marshal(invite)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded JourneyInvite
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if decoded.Type != JourneyInviteType {
		t.Fatalf("Type = %q, want %q", decoded.Type, JourneyInviteType)
	}
	if decoded.Version != JourneyInviteVersion {
		t.Fatalf("Version = %d, want %d", decoded.Version, JourneyInviteVersion)
	}
	if decoded.JourneyID != testJourneyID {
		t.Fatalf("JourneyID = %q, want %q", decoded.JourneyID, testJourneyID)
	}
	if decoded.Audience != JourneyInviteGroupAudience {
		t.Fatalf("Audience = %q, want %q", decoded.Audience, JourneyInviteGroupAudience)
	}
	if decoded.Token.UseMode != JourneyInviteMultiUse {
		t.Fatalf("Token.UseMode = %q, want %q", decoded.Token.UseMode, JourneyInviteMultiUse)
	}
	if decoded.Links == nil || decoded.Links.WebURL == "" || decoded.Links.AppURL == "" {
		t.Fatalf("Links = %#v, want web and app URLs", decoded.Links)
	}
}

func TestJourneyInviteValidateRejectsInvalidFields(t *testing.T) {
	valid := validJourneyInvite(t)

	tests := []struct {
		name   string
		mutate func(*JourneyInvite)
	}{
		{name: "missing invite id", mutate: func(i *JourneyInvite) { i.ID = "" }},
		{name: "missing server url", mutate: func(i *JourneyInvite) { i.ServerURL = "" }},
		{name: "unknown audience", mutate: func(i *JourneyInvite) { i.Audience = JourneyInviteAudience("unknown") }},
		{name: "missing creator membership", mutate: func(i *JourneyInvite) { i.CreatedByJourneyParticipantID = "" }},
		{name: "missing policy hash", mutate: func(i *JourneyInvite) { i.PolicyHash = "" }},
		{name: "missing integrity", mutate: func(i *JourneyInvite) { i.Integrity = nil }},
		{name: "malformed token", mutate: func(i *JourneyInvite) { i.Token.Value = "not base64url!" }},
		{name: "single-use token with too many uses", mutate: func(i *JourneyInvite) {
			i.Token.UseMode = JourneyInviteSingleUse
			i.Token.MaxUses = 2
		}},
		{name: "multi-use token with single-use cap", mutate: func(i *JourneyInvite) {
			i.Token.UseMode = JourneyInviteMultiUse
			i.Token.MaxUses = 1
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invite := valid
			token := valid.Token
			invite.Token = token
			integrity := *valid.Integrity
			invite.Integrity = &integrity
			tt.mutate(&invite)
			if err := invite.Validate(); err == nil {
				t.Fatal("Validate() error = nil, want error")
			}
		})
	}
}

func TestJourneyParticipantValidate(t *testing.T) {
	valid := JourneyParticipant{
		ID:            testMembershipID,
		JourneyID:     testJourneyID,
		ParticipantID: testParticipantID,
		Privileges: JourneyParticipantPrivileges{
			CanGenerateInvites: true,
		},
		JoinedAt: time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC),
	}

	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*JourneyParticipant)
	}{
		{name: "missing membership id", mutate: func(p *JourneyParticipant) { p.ID = "" }},
		{name: "missing journey id", mutate: func(p *JourneyParticipant) { p.JourneyID = "" }},
		{name: "missing participant id", mutate: func(p *JourneyParticipant) { p.ParticipantID = "" }},
		{name: "missing joined at", mutate: func(p *JourneyParticipant) { p.JoinedAt = time.Time{} }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			participant := valid
			tt.mutate(&participant)
			if err := participant.Validate(); err == nil {
				t.Fatal("Validate() error = nil, want error")
			}
		})
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

func validJourney() Journey {
	createdAt := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	activeExpiresAt := time.Date(2026, 5, 24, 18, 0, 0, 0, time.UTC)
	downloadUntil := time.Date(2026, 5, 31, 18, 0, 0, 0, time.UTC)
	purgeAt := time.Date(2026, 6, 1, 18, 0, 0, 0, time.UTC)

	return Journey{
		ID:              testJourneyID,
		OriginServerURL: "https://public.spivot.net",
		Title:           "Sunday Ridge Drive",
		State:           JourneyPlanned,
		Policy: JourneyPolicy{
			PolicyID:        "public-download-7d",
			PolicyHash:      "sha256:abc",
			RetentionMode:   JourneyRetentionDownloadWindow,
			ActiveExpiresAt: &activeExpiresAt,
			DownloadUntil:   &downloadUntil,
			PurgeAt:         &purgeAt,
			ExportSupported: true,
			MediaAllowed:    true,
		},
		Participants: []JourneyParticipant{
			{
				ID:            testMembershipID,
				JourneyID:     testJourneyID,
				ParticipantID: testParticipantID,
				Privileges: JourneyParticipantPrivileges{
					CanGenerateInvites: true,
				},
				JoinedAt: createdAt,
			},
		},
		Segments: []JourneySegment{
			{
				ID:        testSegmentID,
				JourneyID: testJourneyID,
				State:     SegmentPlanned,
				StartedAt: createdAt,
			},
		},
		SharedMedia: []SharedMedia{
			{
				ID:            testMediaID,
				JourneyID:     testJourneyID,
				ParticipantID: testParticipantID,
				ClientAppID:   testClientAppID,
				Type:          MediaPhoto,
				URL:           "https://public.spivot.net/media/99999999-9999-4999-8999-999999999999",
				PolicyHash:    "sha256:abc",
				SharedAt:      createdAt,
			},
		},
		CreatedAt: createdAt,
	}
}

func validJourneyInvite(t *testing.T) JourneyInvite {
	t.Helper()

	expiresAt := time.Date(2026, 5, 24, 12, 30, 0, 0, time.UTC)
	token, err := NewJourneyInviteToken(JourneyInviteMultiUse, expiresAt)
	if err != nil {
		t.Fatalf("NewJourneyInviteToken() error = %v", err)
	}
	token.MaxUses = 10

	invite := NewJourneyInvite("https://public.spivot.net", testJourneyID, token)
	invite.ID = testInviteID
	invite.Audience = JourneyInviteGroupAudience
	invite.CreatedByJourneyParticipantID = testMembershipID
	invite.CreatedAt = time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	invite.PolicyHash = "sha256:abc"
	invite.DisplayName = "Sunday Ridge Drive"
	invite.Links = &JourneyInviteLinks{
		WebURL: "https://public.spivot.net/invites/" + token.Value,
		AppURL: "opencaravan://invite?token=" + token.Value,
	}
	invite.Presentation = &JourneyInvitePresentation{
		Title:   "Join Sunday Ridge Drive",
		Summary: "OpenCaravan group drive invite",
	}
	invite.Integrity = &JourneyInviteIntegrity{
		Algorithm: "ed25519",
		KeyID:     "server-key-1",
		Signature: "base64url-signature",
	}
	return invite
}

func validServerPolicy() ServerPolicy {
	return ServerPolicy{
		ProtocolVersion:        ProtocolVersion,
		ServerURL:              "https://public.spivot.net",
		DisplayName:            "Public Spivot",
		RegistrationMode:       RegistrationOpen,
		DefaultJourneyPolicyID: "public-ephemeral",
		JourneyPolicies: []JourneyPolicyProfile{
			{
				ID:                "public-ephemeral",
				DisplayName:       "Ephemeral",
				Description:       "Hard-purge journey data after the drive ends.",
				RetentionMode:     JourneyRetentionEphemeral,
				MaxActiveLifetime: "24h",
				ExportSupported:   false,
				MediaAllowed:      false,
			},
			{
				ID:                "public-download-7d",
				DisplayName:       "Seven-day download",
				RetentionMode:     JourneyRetentionDownloadWindow,
				MaxActiveLifetime: "24h",
				MaxDownloadWindow: "168h",
				ExportSupported:   true,
				MediaAllowed:      true,
			},
			{
				ID:              "private-forever",
				DisplayName:     "Forever",
				RetentionMode:   JourneyRetentionForever,
				ExportSupported: true,
				MediaAllowed:    true,
			},
		},
	}
}

func ptr[T any](value T) *T {
	return &value
}
