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
	testUserID           UUID = "55555555-5555-4555-8555-555555555555"
	testClientAppID      UUID = "66666666-6666-4666-8666-666666666666"
	testMembershipID     UUID = "77777777-7777-4777-8777-777777777777"
	testInviteID         UUID = "88888888-8888-4888-8888-888888888888"
	testMediaID          UUID = "99999999-9999-4999-8999-999999999999"
	testAvatarImageID    UUID = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	testBannerImageID    UUID = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
	testVehicleAvatarID  UUID = "cccccccc-cccc-4ccc-8ccc-cccccccccccc"

	testUserInactivityDeletionDays int64 = 90
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

func TestHexColorParseAndJSON(t *testing.T) {
	color, err := ParseHexColor("#3366CC")
	if err != nil {
		t.Fatalf("ParseHexColor() error = %v", err)
	}
	if color != HexColor("#3366cc") {
		t.Fatalf("ParseHexColor() = %q, want #3366cc", color)
	}

	encoded, err := json.Marshal(HexColor("#AABBCC"))
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if got, want := string(encoded), `"#aabbcc"`; got != want {
		t.Fatalf("Marshal() = %s, want %s", got, want)
	}

	var decoded HexColor
	if err := json.Unmarshal([]byte(`"#ABCDEF"`), &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if decoded != HexColor("#abcdef") {
		t.Fatalf("decoded color = %q, want #abcdef", decoded)
	}

	for _, value := range []string{"3366cc", "#3366ccff", "#gg66cc", "#12345"} {
		if _, err := ParseHexColor(value); err == nil {
			t.Fatalf("ParseHexColor(%q) error = nil, want error", value)
		}
	}
	if _, err := json.Marshal(HexColor("blue")); err == nil {
		t.Fatal("Marshal() error = nil, want invalid color error")
	}
	if err := json.Unmarshal([]byte(`"blue"`), &decoded); err == nil {
		t.Fatal("Unmarshal() error = nil, want invalid color error")
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
		{name: "open", mode: RegistrationMode("open"), want: false},
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

func TestServerPolicyValidate(t *testing.T) {
	valid := ServerPolicy{
		ProtocolVersion:  ProtocolVersion,
		ServerURL:        "https://public.spivot.net",
		DisplayName:      "Public Spivot",
		RegistrationMode: RegistrationInvite,
		PrivacyURL:       "https://public.spivot.net/privacy",
	}

	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*ServerPolicy)
	}{
		{name: "missing protocol version", mutate: func(p *ServerPolicy) { p.ProtocolVersion = "" }},
		{name: "missing server url", mutate: func(p *ServerPolicy) { p.ServerURL = "" }},
		{name: "missing display name", mutate: func(p *ServerPolicy) { p.DisplayName = "" }},
		{name: "unknown registration mode", mutate: func(p *ServerPolicy) { p.RegistrationMode = RegistrationMode("unknown") }},
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

func TestInviteScopeValid(t *testing.T) {
	tests := []struct {
		name  string
		scope InviteScope
		want  bool
	}{
		{name: "journey", scope: InviteScopeJourney, want: true},
		{name: "server registration", scope: InviteScopeServerRegistration, want: true},
		{name: "unknown", scope: InviteScope("unknown"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.scope.Valid(); got != tt.want {
				t.Fatalf("Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInviteGenerationPermissionsValidate(t *testing.T) {
	valid := validJourneyInviteGenerationPermissions()

	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if err := (InviteGenerationPermissions{}).Validate(); err != nil {
		t.Fatalf("empty Validate() error = %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*InviteGenerationPermissions)
	}{
		{name: "missing scopes", mutate: func(p *InviteGenerationPermissions) { p.Scopes = nil }},
		{name: "invalid scope", mutate: func(p *InviteGenerationPermissions) { p.Scopes[0] = InviteScope("unknown") }},
		{name: "negative redemption cap", mutate: func(p *InviteGenerationPermissions) { p.MaxRedemptionsPerInvite = -1 }},
		{name: "negative lifetime cap", mutate: func(p *InviteGenerationPermissions) { p.MaxLifetimeDays = -1 }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			permissions := valid
			permissions.Scopes = append([]InviteScope(nil), valid.Scopes...)
			tt.mutate(&permissions)
			if err := permissions.Validate(); err == nil {
				t.Fatal("Validate() error = nil, want error")
			}
		})
	}
}

func TestUserProfileContactValidate(t *testing.T) {
	tests := []struct {
		name    string
		contact UserProfileContact
		wantErr bool
	}{
		{
			name: "mobile number",
			contact: UserProfileContact{
				Kind:  UserProfileContactMobileNumber,
				Value: "+15035551212",
			},
		},
		{
			name: "email address",
			contact: UserProfileContact{
				Kind:  UserProfileContactEmailAddress,
				Value: "driver@example.com",
			},
		},
		{
			name: "signal phone link",
			contact: UserProfileContact{
				Kind:  UserProfileContactSignal,
				Value: "https://signal.me/#p/+15035551212",
			},
		},
		{
			name: "signal username link",
			contact: UserProfileContact{
				Kind:  UserProfileContactSignal,
				Value: "https://signal.me/#eu/abcDEF123_-",
			},
		},
		{
			name: "custom kind",
			contact: UserProfileContact{
				Kind:  "club_radio",
				Value: "+15035551212",
			},
		},
		{
			name: "missing kind",
			contact: UserProfileContact{
				Value: "+15035551212",
			},
			wantErr: true,
		},
		{
			name: "missing value",
			contact: UserProfileContact{
				Kind: UserProfileContactMobileNumber,
			},
			wantErr: true,
		},
		{
			name: "invalid mobile number",
			contact: UserProfileContact{
				Kind:  UserProfileContactMobileNumber,
				Value: "503-555-1212",
			},
			wantErr: true,
		},
		{
			name: "invalid email address",
			contact: UserProfileContact{
				Kind:  UserProfileContactEmailAddress,
				Value: "Riley <driver@example.com>",
			},
			wantErr: true,
		},
		{
			name: "signal link must be https",
			contact: UserProfileContact{
				Kind:  UserProfileContactSignal,
				Value: "http://signal.me/#p/+15035551212",
			},
			wantErr: true,
		},
		{
			name: "signal link must use signal.me",
			contact: UserProfileContact{
				Kind:  UserProfileContactSignal,
				Value: "https://example.com/#p/+15035551212",
			},
			wantErr: true,
		},
		{
			name: "signal phone link must contain mobile number",
			contact: UserProfileContact{
				Kind:  UserProfileContactSignal,
				Value: "https://signal.me/#p/503-555-1212",
			},
			wantErr: true,
		},
		{
			name: "signal link must contain fragment",
			contact: UserProfileContact{
				Kind:  UserProfileContactSignal,
				Value: "https://signal.me/",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.contact.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserJSONAndValidate(t *testing.T) {
	user := validUser()

	if err := user.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	encoded, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded User
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if decoded.ID != testUserID {
		t.Fatalf("ID = %q, want %q", decoded.ID, testUserID)
	}
	if decoded.Profile.DisplayName != "Riley" {
		t.Fatalf("DisplayName = %q, want Riley", decoded.Profile.DisplayName)
	}
	if decoded.Profile.AvatarImage == nil || decoded.Profile.AvatarImage.ID != testAvatarImageID {
		t.Fatalf("AvatarImage = %#v, want ID %q", decoded.Profile.AvatarImage, testAvatarImageID)
	}
	if decoded.Profile.AccentColor != HexColor("#3366cc") {
		t.Fatalf("AccentColor = %q, want #3366cc", decoded.Profile.AccentColor)
	}
	if got := decoded.Profile.Contacts[0].Value; got != "+15035551212" {
		t.Fatalf("contact value = %q, want +15035551212", got)
	}
	if decoded.DeletionAfterInactivityDays == nil || *decoded.DeletionAfterInactivityDays != testUserInactivityDeletionDays {
		t.Fatalf("DeletionAfterInactivityDays = %v, want %d", decoded.DeletionAfterInactivityDays, testUserInactivityDeletionDays)
	}
}

func TestUserValidateRejectsInvalidFields(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*User)
	}{
		{name: "missing user id", mutate: func(u *User) { u.ID = "" }},
		{name: "zero inactivity deletion duration", mutate: func(u *User) {
			u.DeletionAfterInactivityDays = ptr[int64](0)
		}},
		{name: "negative inactivity deletion duration", mutate: func(u *User) {
			u.DeletionAfterInactivityDays = ptr[int64](-1)
		}},
		{name: "missing display name", mutate: func(u *User) { u.Profile.DisplayName = "" }},
		{name: "invalid avatar image", mutate: func(u *User) { u.Profile.AvatarImage.ID = "" }},
		{name: "invalid banner image", mutate: func(u *User) { u.Profile.BannerImage.ContentType = "text/plain" }},
		{name: "invalid accent color", mutate: func(u *User) { u.Profile.AccentColor = "blue" }},
		{name: "invalid permissions", mutate: func(u *User) { u.Permissions.InviteGeneration.Scopes[0] = InviteScope("unknown") }},
		{name: "invalid profile link", mutate: func(u *User) { u.Profile.Links[0].URL = "/relative" }},
		{name: "missing contact kind", mutate: func(u *User) { u.Profile.Contacts[0].Kind = "" }},
		{name: "invalid mobile contact", mutate: func(u *User) { u.Profile.Contacts[0].Value = "503-555-1212" }},
		{name: "client app for other user", mutate: func(u *User) { u.ClientApps[0].UserID = testVehicleID }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := validUser()
			tt.mutate(&user)
			if err := user.Validate(); err == nil {
				t.Fatal("Validate() error = nil, want error")
			}
		})
	}
}

func TestImageResourceRefValidate(t *testing.T) {
	valid := validAvatarImageRef()

	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*ImageResourceRef)
	}{
		{name: "missing id", mutate: func(ref *ImageResourceRef) { ref.ID = "" }},
		{name: "missing digest", mutate: func(ref *ImageResourceRef) { ref.Digest = "" }},
		{name: "not image content type", mutate: func(ref *ImageResourceRef) { ref.ContentType = "application/octet-stream" }},
		{name: "negative width", mutate: func(ref *ImageResourceRef) { ref.WidthPixels = -1 }},
		{name: "negative height", mutate: func(ref *ImageResourceRef) { ref.HeightPixels = -1 }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := valid
			tt.mutate(&ref)
			if err := ref.Validate(); err == nil {
				t.Fatal("Validate() error = nil, want error")
			}
		})
	}
}

func TestVehicleValidate(t *testing.T) {
	valid := validVehicle()

	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*Vehicle)
	}{
		{name: "missing vehicle id", mutate: func(vehicle *Vehicle) { vehicle.ID = "" }},
		{name: "missing display name", mutate: func(vehicle *Vehicle) { vehicle.DisplayName = "" }},
		{name: "invalid avatar image", mutate: func(vehicle *Vehicle) { vehicle.AvatarImage.Digest = "" }},
		{name: "invalid banner image", mutate: func(vehicle *Vehicle) { vehicle.BannerImage.ContentType = "application/json" }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vehicle := validVehicle()
			tt.mutate(&vehicle)
			if err := vehicle.Validate(); err == nil {
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

	forever := valid
	forever.DeletionTime = nil
	if err := forever.Validate(); err != nil {
		t.Fatalf("forever Validate() error = %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*Journey)
	}{
		{name: "missing journey id", mutate: func(j *Journey) { j.ID = "" }},
		{name: "missing origin server", mutate: func(j *Journey) { j.OriginServerURL = "" }},
		{name: "missing title", mutate: func(j *Journey) { j.Title = "" }},
		{name: "unknown state", mutate: func(j *Journey) { j.State = JourneyState("unknown") }},
		{name: "invalid avatar image", mutate: func(j *Journey) {
			invalid := validAvatarImageRef()
			invalid.ID = ""
			j.AvatarImage = &invalid
		}},
		{name: "invalid banner image", mutate: func(j *Journey) {
			invalid := validBannerImageRef()
			invalid.ContentType = "application/json"
			j.BannerImage = &invalid
		}},
		{name: "zero deletion time", mutate: func(j *Journey) { j.DeletionTime = ptr(time.Time{}) }},
		{name: "missing creation time", mutate: func(j *Journey) { j.CreationTime = time.Time{} }},
		{name: "deletion before creation", mutate: func(j *Journey) {
			beforeCreation := j.CreationTime.Add(-time.Minute)
			j.DeletionTime = &beforeCreation
		}},
		{name: "zero planned start time", mutate: func(j *Journey) { j.PlannedStartTime = ptr(time.Time{}) }},
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

func TestNewInviteToken(t *testing.T) {
	expirationTime := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)

	token, err := NewInviteToken(expirationTime)
	if err != nil {
		t.Fatalf("NewInviteToken() error = %v", err)
	}
	tokenBytes, err := base64.RawURLEncoding.DecodeString(token.Value)
	if err != nil {
		t.Fatalf("DecodeString() error = %v", err)
	}
	if len(tokenBytes) != InviteTokenBytes {
		t.Fatalf("decoded token length = %d, want %d", len(tokenBytes), InviteTokenBytes)
	}

	if _, err := NewInviteToken(time.Time{}); err == nil {
		t.Fatal("NewInviteToken() error = nil, want missing expiration_time error")
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
	if decoded.MaxRedemptions != 10 {
		t.Fatalf("MaxRedemptions = %d, want 10", decoded.MaxRedemptions)
	}
	if decoded.Links == nil || decoded.Links.WebURL == "" || decoded.Links.AppURL == "" {
		t.Fatalf("Links = %#v, want web and app URLs", decoded.Links)
	}
	if decoded.Presentation == nil || decoded.Presentation.AvatarImage == nil ||
		decoded.Presentation.AvatarImage.ID != testAvatarImageID {
		t.Fatalf("Presentation.AvatarImage = %#v, want ID %q", decoded.Presentation, testAvatarImageID)
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
		{name: "missing creator membership", mutate: func(i *JourneyInvite) { i.CreatedByJourneyParticipantID = "" }},
		{name: "missing policy hash", mutate: func(i *JourneyInvite) { i.PolicyHash = "" }},
		{name: "missing integrity", mutate: func(i *JourneyInvite) { i.Integrity = nil }},
		{name: "malformed token", mutate: func(i *JourneyInvite) { i.Token.Value = "not base64url!" }},
		{name: "negative max redemptions", mutate: func(i *JourneyInvite) { i.MaxRedemptions = -1 }},
		{name: "invalid presentation avatar image", mutate: func(i *JourneyInvite) {
			invalid := validAvatarImageRef()
			invalid.Digest = ""
			i.Presentation = &JourneyInvitePresentation{AvatarImage: &invalid}
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
		ID:        testMembershipID,
		JourneyID: testJourneyID,
		UserID:    testUserID,
		Privileges: JourneyParticipantPrivileges{
			InviteGeneration: ptr(validJourneyInviteGenerationPermissions()),
		},
		JoinTime: time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC),
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
		{name: "missing user id", mutate: func(p *JourneyParticipant) { p.UserID = "" }},
		{name: "invalid profile", mutate: func(p *JourneyParticipant) { p.Profile = &UserProfile{} }},
		{name: "invalid privileges", mutate: func(p *JourneyParticipant) { p.Privileges.InviteGeneration.Scopes[0] = InviteScope("unknown") }},
		{name: "missing join time", mutate: func(p *JourneyParticipant) { p.JoinTime = time.Time{} }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			participant := valid
			if valid.Privileges.InviteGeneration != nil {
				permissions := *valid.Privileges.InviteGeneration
				permissions.Scopes = append([]InviteScope(nil), permissions.Scopes...)
				participant.Privileges.InviteGeneration = &permissions
			}
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
		{name: "missing journey participant id", mutate: func(s *PositionSample) { s.JourneyParticipantID = "" }},
		{name: "missing client app id", mutate: func(s *PositionSample) { s.ClientAppID = "" }},
		{name: "negative sequence", mutate: func(s *PositionSample) { s.ClientSequence = -1 }},
		{name: "missing capture time", mutate: func(s *PositionSample) { s.CaptureTime = time.Time{} }},
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
				JourneyParticipantID: testMembershipID,
				ClientAppIDs:         []UUID{testClientAppID},
				Role:                 OccupantDriver,
				JoinTime:             time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC),
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
		{name: "tracklog from non occupant", mutate: func(v *SegmentVehicle) { v.Tracklog[0].JourneyParticipantID = testVehicleID }},
		{name: "tracklog for other segment vehicle", mutate: func(v *SegmentVehicle) { v.Tracklog[0].SegmentVehicleID = testVehicleID }},
		{name: "tracklog for other segment", mutate: func(v *SegmentVehicle) { v.Tracklog[0].SegmentID = testVehicleID }},
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
		JourneyID:            testJourneyID,
		SegmentID:            testSegmentID,
		SegmentVehicleID:     testSegmentVehicleID,
		JourneyParticipantID: testMembershipID,
		ClientAppID:          testClientAppID,
		ClientSequence:       1,
		CaptureTime:          time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC),
		LatitudeE7:           451234567,
		LongitudeE7:          -1221234567,
		HeadingDegreesE2:     ptr[int32](35999),
	}
}

func validUser() User {
	return User{
		ID: testUserID,
		Permissions: &UserPermissions{
			InviteGeneration: ptr(InviteGenerationPermissions{
				Scopes:                  []InviteScope{InviteScopeServerRegistration},
				MaxLifetimeDays:         30,
				MaxRedemptionsPerInvite: 1,
			}),
		},
		Profile: UserProfile{
			DisplayName: "Riley",
			AvatarImage: ptr(validAvatarImageRef()),
			BannerImage: ptr(validBannerImageRef()),
			Bio:         "Usually somewhere near the back of the convoy.",
			AccentColor: "#3366CC",
			Links: []UserProfileLink{
				{
					Kind:  "website",
					Label: "Road notes",
					URL:   "https://example.com/riley",
				},
			},
			Contacts: []UserProfileContact{
				{
					Kind:        UserProfileContactMobileNumber,
					Label:       "Text me",
					DisplayText: "+1 503 555 1212",
					Value:       "+15035551212",
					Verified:    true,
				},
				{
					Kind:  UserProfileContactSignal,
					Label: "Signal",
					Value: "https://signal.me/#p/+15035551212",
				},
			},
		},
		DeletionAfterInactivityDays: ptr(testUserInactivityDeletionDays),
		ClientApps: []ClientApp{
			{
				ID:       testClientAppID,
				UserID:   testUserID,
				Name:     "Spivot",
				Version:  "0.1.0",
				Platform: "ios",
			},
		},
	}
}

func validVehicle() Vehicle {
	return Vehicle{
		ID:          testVehicleID,
		DisplayName: "Blue Bronco",
		Make:        "Ford",
		Model:       "Bronco",
		ModelYear:   2026,
		Color:       "blue",
		AvatarImage: ptr(validVehicleAvatarImageRef()),
		BannerImage: ptr(validBannerImageRef()),
	}
}

func validJourneyInviteGenerationPermissions() InviteGenerationPermissions {
	return InviteGenerationPermissions{
		Scopes:                  []InviteScope{InviteScopeJourney},
		MaxRedemptionsPerInvite: 25,
		MaxLifetimeDays:         7,
	}
}

func validAvatarImageRef() ImageResourceRef {
	return ImageResourceRef{
		ID:           testAvatarImageID,
		Digest:       "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		ContentType:  "image/png",
		WidthPixels:  512,
		HeightPixels: 512,
	}
}

func validVehicleAvatarImageRef() ImageResourceRef {
	return ImageResourceRef{
		ID:           testVehicleAvatarID,
		Digest:       "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
		ContentType:  "image/png",
		WidthPixels:  512,
		HeightPixels: 512,
	}
}

func validBannerImageRef() ImageResourceRef {
	return ImageResourceRef{
		ID:           testBannerImageID,
		Digest:       "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		ContentType:  "image/jpeg",
		WidthPixels:  1200,
		HeightPixels: 400,
	}
}

func validJourney() Journey {
	creationTime := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	deletionTime := time.Date(2026, 6, 1, 18, 0, 0, 0, time.UTC)
	profile := validUser().Profile

	return Journey{
		ID:              testJourneyID,
		OriginServerURL: "https://public.spivot.net",
		Title:           "Sunday Ridge Drive",
		AvatarImage:     ptr(validAvatarImageRef()),
		BannerImage:     ptr(validBannerImageRef()),
		State:           JourneyPlanned,
		DeletionTime:    &deletionTime,
		Features: JourneyFeatures{
			ExportAllowed: true,
			MediaAllowed:  true,
		},
		Participants: []JourneyParticipant{
			{
				ID:        testMembershipID,
				JourneyID: testJourneyID,
				UserID:    testUserID,
				Profile:   &profile,
				Privileges: JourneyParticipantPrivileges{
					InviteGeneration: ptr(validJourneyInviteGenerationPermissions()),
				},
				JoinTime: creationTime,
			},
		},
		Segments: []JourneySegment{
			{
				ID:        testSegmentID,
				JourneyID: testJourneyID,
				State:     SegmentPlanned,
				StartTime: creationTime,
			},
		},
		SharedMedia: []SharedMedia{
			{
				ID:                   testMediaID,
				JourneyID:            testJourneyID,
				JourneyParticipantID: testMembershipID,
				ClientAppID:          testClientAppID,
				Type:                 MediaPhoto,
				URL:                  "https://public.spivot.net/media/99999999-9999-4999-8999-999999999999",
				PolicyHash:           "sha256:abc",
				ShareTime:            creationTime,
			},
		},
		CreationTime: creationTime,
	}
}

func validJourneyInvite(t *testing.T) JourneyInvite {
	t.Helper()

	expirationTime := time.Date(2026, 5, 24, 12, 30, 0, 0, time.UTC)
	token, err := NewInviteToken(expirationTime)
	if err != nil {
		t.Fatalf("NewInviteToken() error = %v", err)
	}

	invite := NewJourneyInvite("https://public.spivot.net", testJourneyID, token, 10)
	invite.ID = testInviteID
	invite.CreatedByJourneyParticipantID = testMembershipID
	invite.CreationTime = time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	invite.PolicyHash = "sha256:abc"
	invite.DisplayName = "Sunday Ridge Drive"
	invite.Links = &JourneyInviteLinks{
		WebURL: "https://public.spivot.net/invites/" + token.Value,
		AppURL: "opencaravan://invite?token=" + token.Value,
	}
	invite.Presentation = &JourneyInvitePresentation{
		Title:       "Join Sunday Ridge Drive",
		Summary:     "OpenCaravan journey invite",
		AvatarImage: ptr(validAvatarImageRef()),
		BannerImage: ptr(validBannerImageRef()),
	}
	invite.Integrity = &Integrity{
		Algorithm: "ed25519",
		KeyID:     "server-key-1",
		Signature: "base64url-signature",
	}
	return invite
}

func ptr[T any](value T) *T {
	return &value
}
