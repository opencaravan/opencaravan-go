package opencaravan_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	opencaravan "github.com/opencaravan/opencaravan-go"
)

// mustUUID returns a fresh UUID that passes UUID.Valid, failing the test on
// any underlying generator error. Used so every fixture can have unique
// identity-bearing fields without forcing the call sites to handle the
// (vanishingly unlikely) UUID generation failure path.
func mustUUID(t *testing.T) opencaravan.UUID {
	t.Helper()
	id, err := opencaravan.NewUUID()
	if err != nil {
		t.Fatalf("NewUUID: %v", err)
	}
	return id
}

func validVehicle(t *testing.T) opencaravan.Vehicle {
	t.Helper()
	return opencaravan.Vehicle{
		ID:                mustUUID(t),
		DisplayName:       "Riley's Subaru",
		Make:              "Subaru",
		Model:             "Outback",
		ModelYear:         2018,
		Color:             "silver",
		OwnerUserID:       mustUUID(t),
		Capacity:          5,
		AuthorizedDrivers: []opencaravan.UUID{mustUUID(t)},
		ACLVersion:        1,
	}
}

func TestCanonicalJSONDeterministic(t *testing.T) {
	v := validVehicle(t)
	first, err := opencaravan.CanonicalJSON(v)
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	for i := 0; i < 5; i++ {
		got, err := opencaravan.CanonicalJSON(v)
		if err != nil {
			t.Fatalf("iter %d: %v", i, err)
		}
		if string(got) != string(first) {
			t.Fatalf("iter %d differs:\n  first: %s\n  got:   %s", i, first, got)
		}
	}
}

func TestCanonicalJSONSortsKeysAndOmitsWhitespace(t *testing.T) {
	v := validVehicle(t)
	got, err := opencaravan.CanonicalJSON(v)
	if err != nil {
		t.Fatalf("canonical: %v", err)
	}
	s := string(got)
	// The canonical encoder must not emit insignificant whitespace between
	// JSON tokens. String literal contents may of course contain spaces
	// (display_name = "Riley's Subaru"), so check for the indent patterns
	// json.MarshalIndent would introduce — colon-space and comma-space —
	// rather than any whitespace character.
	for _, pattern := range []string{`": `, `, "`} {
		if strings.Contains(s, pattern) {
			t.Fatalf("canonical output contains insignificant whitespace %q in: %s", pattern, s)
		}
	}
	// Verify lexicographic ordering: "acl_version" comes before "authorized_drivers",
	// "authorized_drivers" before "capacity", "capacity" before "color", and so on.
	// Confirm by checking a few representative pairs.
	for _, pair := range []struct{ before, after string }{
		{`"acl_version"`, `"authorized_drivers"`},
		{`"authorized_drivers"`, `"capacity"`},
		{`"capacity"`, `"color"`},
		{`"display_name"`, `"id"`},
		{`"id"`, `"make"`},
		{`"make"`, `"model"`},
		{`"model"`, `"model_year"`},
		{`"model_year"`, `"owner_user_id"`},
	} {
		b := strings.Index(s, pair.before)
		a := strings.Index(s, pair.after)
		if b < 0 || a < 0 {
			t.Fatalf("expected both %s and %s in: %s", pair.before, pair.after, s)
		}
		if b >= a {
			t.Fatalf("%s should precede %s; got positions %d and %d in: %s", pair.before, pair.after, b, a, s)
		}
	}
}

func TestCanonicalJSONOmitsEmptyOptionalFields(t *testing.T) {
	// A Vehicle without optional fields (avatar_image, banner_image,
	// emergency_rule, integrity, make, model, etc.) should produce canonical
	// JSON that does NOT contain those keys.
	v := opencaravan.Vehicle{
		ID:                mustUUID(t),
		DisplayName:       "Spare",
		OwnerUserID:       mustUUID(t),
		Capacity:          4,
		AuthorizedDrivers: []opencaravan.UUID{mustUUID(t)},
		ACLVersion:        1,
	}
	got, err := opencaravan.CanonicalJSON(v)
	if err != nil {
		t.Fatalf("canonical: %v", err)
	}
	s := string(got)
	for _, omitted := range []string{
		`"avatar_image"`, `"banner_image"`, `"make"`, `"model"`,
		`"model_year"`, `"color"`, `"emergency_rule"`, `"integrity"`,
	} {
		if strings.Contains(s, omitted) {
			t.Fatalf("expected %s to be omitted from canonical encoding; got: %s", omitted, s)
		}
	}
}

func TestVehicleCanonicalEncodingExcludesIntegrity(t *testing.T) {
	v := validVehicle(t)
	signedBytes, err := v.CanonicalEncoding()
	if err != nil {
		t.Fatalf("canonical before: %v", err)
	}

	// Attach an Integrity envelope and re-canonicalize. The bytes must be
	// identical — the signature cannot cover itself, so adding it must not
	// change the input the signature is computed over.
	v.Integrity = &opencaravan.Integrity{
		Algorithm: "p256-ecdsa-sha256",
		KeyID:     "sha256:abcdef",
		Signature: base64.RawURLEncoding.EncodeToString([]byte("not-a-real-sig")),
	}
	signedBytesAfter, err := v.CanonicalEncoding()
	if err != nil {
		t.Fatalf("canonical after: %v", err)
	}
	if string(signedBytes) != string(signedBytesAfter) {
		t.Fatalf("CanonicalEncoding must exclude Integrity\n  before: %s\n  after:  %s", signedBytes, signedBytesAfter)
	}
}

func TestVehicleValidate(t *testing.T) {
	good := validVehicle(t)
	if err := good.Validate(); err != nil {
		t.Fatalf("happy path: %v", err)
	}

	cases := map[string]func(*opencaravan.Vehicle){
		"missing id":             func(v *opencaravan.Vehicle) { v.ID = "" },
		"missing display_name":   func(v *opencaravan.Vehicle) { v.DisplayName = "" },
		"missing owner_user_id":  func(v *opencaravan.Vehicle) { v.OwnerUserID = "" },
		"zero capacity":          func(v *opencaravan.Vehicle) { v.Capacity = 0 },
		"negative capacity":      func(v *opencaravan.Vehicle) { v.Capacity = -1 },
		"zero acl_version":       func(v *opencaravan.Vehicle) { v.ACLVersion = 0 },
		"bad authorized driver":  func(v *opencaravan.Vehicle) { v.AuthorizedDrivers = []opencaravan.UUID{"not-a-uuid"} },
		"bad emergency rule":     func(v *opencaravan.Vehicle) { v.EmergencyRule = &opencaravan.VehicleEmergencyRule{Kind: "unknown"} },
		"bad integrity envelope": func(v *opencaravan.Vehicle) { v.Integrity = &opencaravan.Integrity{} },
	}
	for name, mut := range cases {
		t.Run(name, func(t *testing.T) {
			v := validVehicle(t)
			mut(&v)
			if err := v.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func validACL(t *testing.T) opencaravan.VehicleACL {
	t.Helper()
	return opencaravan.VehicleACL{
		VehicleID:         mustUUID(t),
		OwnerUserID:       mustUUID(t),
		ACLVersion:        2,
		AuthorizedDrivers: []opencaravan.UUID{mustUUID(t), mustUUID(t)},
		EffectiveTime:     time.Now().UTC(),
	}
}

func TestVehicleACLValidate(t *testing.T) {
	good := validACL(t)
	if err := good.Validate(); err != nil {
		t.Fatalf("happy path: %v", err)
	}
	cases := map[string]func(*opencaravan.VehicleACL){
		"missing vehicle_id":    func(a *opencaravan.VehicleACL) { a.VehicleID = "" },
		"missing owner_user_id": func(a *opencaravan.VehicleACL) { a.OwnerUserID = "" },
		"zero acl_version":      func(a *opencaravan.VehicleACL) { a.ACLVersion = 0 },
		"bad driver uuid":       func(a *opencaravan.VehicleACL) { a.AuthorizedDrivers = []opencaravan.UUID{"x"} },
		"missing effective":     func(a *opencaravan.VehicleACL) { a.EffectiveTime = time.Time{} },
		"bad emergency rule":    func(a *opencaravan.VehicleACL) { a.EmergencyRule = &opencaravan.VehicleEmergencyRule{Kind: "x"} },
	}
	for name, mut := range cases {
		t.Run(name, func(t *testing.T) {
			a := validACL(t)
			mut(&a)
			if err := a.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func validAttestation(t *testing.T) opencaravan.DriverAttestation {
	t.Helper()
	return opencaravan.DriverAttestation{
		VehicleID:           mustUUID(t),
		SegmentID:           mustUUID(t),
		DriverUserID:        mustUUID(t),
		EffectiveTime:       time.Now().UTC(),
		ACLVersionConsulted: 1,
	}
}

func TestDriverAttestationValidate(t *testing.T) {
	good := validAttestation(t)
	if err := good.Validate(); err != nil {
		t.Fatalf("happy path: %v", err)
	}
	withChain := validAttestation(t)
	hash := "sha256:" + strings.Repeat("a", 64)
	withChain.PriorAttestationHash = &hash
	if err := withChain.Validate(); err != nil {
		t.Fatalf("with valid prior chain: %v", err)
	}

	cases := map[string]func(*opencaravan.DriverAttestation){
		"missing vehicle_id":     func(a *opencaravan.DriverAttestation) { a.VehicleID = "" },
		"missing segment_id":     func(a *opencaravan.DriverAttestation) { a.SegmentID = "" },
		"missing driver_user_id": func(a *opencaravan.DriverAttestation) { a.DriverUserID = "" },
		"missing effective_time": func(a *opencaravan.DriverAttestation) { a.EffectiveTime = time.Time{} },
		"zero acl_version":       func(a *opencaravan.DriverAttestation) { a.ACLVersionConsulted = 0 },
		"empty prior hash": func(a *opencaravan.DriverAttestation) {
			s := ""
			a.PriorAttestationHash = &s
		},
		"prior hash missing sha256 prefix": func(a *opencaravan.DriverAttestation) {
			s := "md5:" + strings.Repeat("a", 64)
			a.PriorAttestationHash = &s
		},
		"prior hash hex too short": func(a *opencaravan.DriverAttestation) {
			s := "sha256:abc"
			a.PriorAttestationHash = &s
		},
		"prior hash hex too long": func(a *opencaravan.DriverAttestation) {
			s := "sha256:" + strings.Repeat("a", 65)
			a.PriorAttestationHash = &s
		},
		"prior hash uppercase hex": func(a *opencaravan.DriverAttestation) {
			s := "sha256:" + strings.Repeat("A", 64)
			a.PriorAttestationHash = &s
		},
		"prior hash non-hex char": func(a *opencaravan.DriverAttestation) {
			s := "sha256:" + strings.Repeat("g", 64)
			a.PriorAttestationHash = &s
		},
	}
	for name, mut := range cases {
		t.Run(name, func(t *testing.T) {
			a := validAttestation(t)
			mut(&a)
			if err := a.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestDriverAttestationChainHashIncludedInSignedBytes(t *testing.T) {
	// The PriorAttestationHash IS part of the signed input (the comment on
	// CanonicalEncoding promises this). Two attestations identical in every
	// field except PriorAttestationHash must produce different canonical
	// bytes — otherwise an attacker could swap chain hashes after signing.
	base := validAttestation(t)
	withoutChain, err := base.CanonicalEncoding()
	if err != nil {
		t.Fatalf("encode without chain: %v", err)
	}
	chained := base
	hash := "sha256:" + strings.Repeat("b", 64)
	chained.PriorAttestationHash = &hash
	withChain, err := chained.CanonicalEncoding()
	if err != nil {
		t.Fatalf("encode with chain: %v", err)
	}
	if string(withoutChain) == string(withChain) {
		t.Fatalf("PriorAttestationHash must contribute to signed bytes; both encodings equal:\n%s", withoutChain)
	}
}

func TestVehicleEmergencyRuleKindValid(t *testing.T) {
	for _, k := range []opencaravan.VehicleEmergencyRuleKind{
		opencaravan.VehicleEmergencyRuleNone,
		opencaravan.VehicleEmergencyRuleAnyJourneyParticipant,
	} {
		if !k.Valid() {
			t.Errorf("%q expected Valid()=true", k)
		}
	}
	for _, k := range []opencaravan.VehicleEmergencyRuleKind{
		"", "anyone", "magic",
	} {
		if k.Valid() {
			t.Errorf("%q expected Valid()=false", k)
		}
	}
}

func TestVehicleSignVerifyRoundTrip(t *testing.T) {
	// End-to-end sign + verify on real bytes: produce CanonicalEncoding,
	// sign with a fresh P-256 key, verify against the canonical bytes,
	// confirm validity. Then mutate one byte and confirm verification
	// fails. This is the property a conformant implementation must support
	// in another language.
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}
	v := validVehicle(t)
	signedBytes, err := v.CanonicalEncoding()
	if err != nil {
		t.Fatalf("canonical: %v", err)
	}
	digest := sha256.Sum256(signedBytes)
	sig, err := ecdsa.SignASN1(rand.Reader, priv, digest[:])
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	if !ecdsa.VerifyASN1(&priv.PublicKey, digest[:], sig) {
		t.Fatal("verify of fresh signature failed")
	}

	// Attach integrity, re-derive canonical bytes (should be identical), verify again.
	v.Integrity = &opencaravan.Integrity{
		Algorithm: "p256-ecdsa-sha256",
		KeyID:     "sha256:" + strings.Repeat("c", 64),
		Signature: base64.RawURLEncoding.EncodeToString(sig),
	}
	after, err := v.CanonicalEncoding()
	if err != nil {
		t.Fatalf("canonical after integrity: %v", err)
	}
	if string(after) != string(signedBytes) {
		t.Fatal("CanonicalEncoding must be stable with or without Integrity")
	}

	// Mutate one byte in the canonical input — verification must fail.
	mutated := make([]byte, len(signedBytes))
	copy(mutated, signedBytes)
	mutated[0] ^= 0x01
	mutatedDigest := sha256.Sum256(mutated)
	if ecdsa.VerifyASN1(&priv.PublicKey, mutatedDigest[:], sig) {
		t.Fatal("verify of mutated input must fail")
	}
}

func TestVehicleJSONRoundTrip(t *testing.T) {
	// Round-trip through JSON: a Vehicle marshal/unmarshal must be lossless
	// for every field, including the new auth metadata.
	v := validVehicle(t)
	v.AvatarImage = &opencaravan.ImageResourceRef{
		ID: mustUUID(t), Digest: "sha256:" + strings.Repeat("d", 64),
		ContentType: "image/png", WidthPixels: 512, HeightPixels: 512,
	}
	v.EmergencyRule = &opencaravan.VehicleEmergencyRule{
		Kind: opencaravan.VehicleEmergencyRuleAnyJourneyParticipant,
	}
	v.Integrity = &opencaravan.Integrity{
		Algorithm: "p256-ecdsa-sha256",
		KeyID:     "sha256:" + strings.Repeat("e", 64),
		Signature: base64.RawURLEncoding.EncodeToString([]byte("sig")),
	}
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got opencaravan.Vehicle
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.ID != v.ID || got.OwnerUserID != v.OwnerUserID ||
		got.Capacity != v.Capacity || got.ACLVersion != v.ACLVersion ||
		got.EmergencyRule == nil || got.EmergencyRule.Kind != v.EmergencyRule.Kind ||
		got.Integrity == nil || got.Integrity.Algorithm != v.Integrity.Algorithm {
		t.Fatalf("round-trip lossy:\n  want: %+v\n  got:  %+v", v, got)
	}
}
