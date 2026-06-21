package opencaravan_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"strings"
	"testing"
	"time"

	opencaravan "github.com/opencaravan/opencaravan-go"
)

func validGarage(t *testing.T) opencaravan.Garage {
	t.Helper()
	owner := mustUUID(t)
	now := time.Now().UTC().Truncate(time.Nanosecond)
	accepted := now
	return opencaravan.Garage{
		ID:              mustUUID(t),
		Name:            "Wheelsdown Household",
		RevisionVersion: 1,
		RevisionTime:    now,
		Owners: []opencaravan.GarageOwner{
			{UserID: owner, AddedTime: now, AcceptedTime: &accepted},
		},
		SignedBy: owner,
	}
}

func TestGarageValidate(t *testing.T) {
	good := validGarage(t)
	if err := good.Validate(); err != nil {
		t.Fatalf("happy path: %v", err)
	}

	cases := map[string]func(*opencaravan.Garage){
		"missing id":            func(g *opencaravan.Garage) { g.ID = "" },
		"missing name":          func(g *opencaravan.Garage) { g.Name = "" },
		"zero revision":         func(g *opencaravan.Garage) { g.RevisionVersion = 0 },
		"missing revision time": func(g *opencaravan.Garage) { g.RevisionTime = time.Time{} },
		"no owners":             func(g *opencaravan.Garage) { g.Owners = nil },
		"signer not in owners": func(g *opencaravan.Garage) {
			other, _ := opencaravan.NewUUID()
			g.SignedBy = other
		},
		"signer pending acceptance": func(g *opencaravan.Garage) {
			g.Owners[0].AcceptedTime = nil
		},
		"duplicate owner uuid": func(g *opencaravan.Garage) {
			now := g.RevisionTime
			accepted := now
			g.Owners = append(g.Owners, opencaravan.GarageOwner{
				UserID: g.Owners[0].UserID, AddedTime: now, AcceptedTime: &accepted,
			})
		},
		"bad integrity envelope": func(g *opencaravan.Garage) {
			g.Integrity = &opencaravan.Integrity{}
		},
	}
	for name, mut := range cases {
		t.Run(name, func(t *testing.T) {
			g := validGarage(t)
			mut(&g)
			if err := g.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestGarageOwnerValidate(t *testing.T) {
	now := time.Now().UTC()
	good := opencaravan.GarageOwner{UserID: mustUUID(t), AddedTime: now}
	if err := good.Validate(); err != nil {
		t.Fatalf("pending owner: %v", err)
	}
	accepted := now.Add(time.Minute)
	acceptedOwner := opencaravan.GarageOwner{UserID: mustUUID(t), AddedTime: now, AcceptedTime: &accepted}
	if err := acceptedOwner.Validate(); err != nil {
		t.Fatalf("accepted owner: %v", err)
	}

	// accepted_time before added_time is structurally invalid (no time travel).
	earlier := now.Add(-time.Minute)
	timeTravel := opencaravan.GarageOwner{UserID: mustUUID(t), AddedTime: now, AcceptedTime: &earlier}
	if err := timeTravel.Validate(); err == nil {
		t.Fatal("expected rejection of accepted_time before added_time")
	}

	// zero accepted_time-when-set is invalid (pending = nil, not zero value).
	var zeroT time.Time
	bogus := opencaravan.GarageOwner{UserID: mustUUID(t), AddedTime: now, AcceptedTime: &zeroT}
	if err := bogus.Validate(); err == nil {
		t.Fatal("expected rejection of zero accepted_time")
	}
}

func TestGarageCanonicalEncodingExcludesIntegrity(t *testing.T) {
	g := validGarage(t)
	before, err := g.CanonicalEncoding()
	if err != nil {
		t.Fatalf("encode before: %v", err)
	}
	g.Integrity = &opencaravan.Integrity{
		Algorithm: "p256-ecdsa-sha256",
		KeyID:     "sha256:" + strings.Repeat("a", 64),
		Signature: base64.RawURLEncoding.EncodeToString([]byte("sig")),
	}
	after, err := g.CanonicalEncoding()
	if err != nil {
		t.Fatalf("encode after: %v", err)
	}
	if string(before) != string(after) {
		t.Fatal("CanonicalEncoding must exclude Integrity")
	}
}

func TestGarageSignVerifyRoundTrip(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}
	g := validGarage(t)
	signedBytes, err := g.CanonicalEncoding()
	if err != nil {
		t.Fatalf("canonical: %v", err)
	}
	digest := sha256.Sum256(signedBytes)
	sig, err := ecdsa.SignASN1(rand.Reader, priv, digest[:])
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if !ecdsa.VerifyASN1(&priv.PublicKey, digest[:], sig) {
		t.Fatal("fresh signature verify failed")
	}
	mutated := make([]byte, len(signedBytes))
	copy(mutated, signedBytes)
	mutated[len(mutated)-2] ^= 0x01
	mutatedDigest := sha256.Sum256(mutated)
	if ecdsa.VerifyASN1(&priv.PublicKey, mutatedDigest[:], sig) {
		t.Fatal("verify of mutated input must fail")
	}
}

func validAcceptance(t *testing.T) opencaravan.GarageOwnershipAcceptance {
	t.Helper()
	return opencaravan.GarageOwnershipAcceptance{
		GarageID:                mustUUID(t),
		RevisionVersionAccepted: 2,
		AccepterUserID:          mustUUID(t),
		AcceptedTime:            time.Now().UTC(),
	}
}

func TestGarageOwnershipAcceptanceValidate(t *testing.T) {
	good := validAcceptance(t)
	if err := good.Validate(); err != nil {
		t.Fatalf("happy path: %v", err)
	}
	cases := map[string]func(*opencaravan.GarageOwnershipAcceptance){
		"missing garage_id":        func(a *opencaravan.GarageOwnershipAcceptance) { a.GarageID = "" },
		"zero revision":            func(a *opencaravan.GarageOwnershipAcceptance) { a.RevisionVersionAccepted = 0 },
		"missing accepter_user_id": func(a *opencaravan.GarageOwnershipAcceptance) { a.AccepterUserID = "" },
		"missing accepted_time":    func(a *opencaravan.GarageOwnershipAcceptance) { a.AcceptedTime = time.Time{} },
		"bad integrity envelope":   func(a *opencaravan.GarageOwnershipAcceptance) { a.Integrity = &opencaravan.Integrity{} },
	}
	for name, mut := range cases {
		t.Run(name, func(t *testing.T) {
			a := validAcceptance(t)
			mut(&a)
			if err := a.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func validGarageVehicle(t *testing.T) opencaravan.GarageVehicle {
	t.Helper()
	return opencaravan.GarageVehicle{
		ID:              mustUUID(t),
		GarageID:        mustUUID(t),
		RevisionVersion: 1,
		RevisionTime:    time.Now().UTC(),
		DisplayName:     "Riley's Subaru",
		Make:            "Subaru",
		Model:           "Outback",
		ModelYear:       2018,
		Color:           "silver",
		Capacity:        5,
		Notes:           "Transmission rebuilt 2024-03.",
		SignedBy:        mustUUID(t),
	}
}

func TestGarageVehicleValidate(t *testing.T) {
	good := validGarageVehicle(t)
	if err := good.Validate(); err != nil {
		t.Fatalf("happy path: %v", err)
	}
	cases := map[string]func(*opencaravan.GarageVehicle){
		"missing id":            func(gv *opencaravan.GarageVehicle) { gv.ID = "" },
		"missing garage_id":     func(gv *opencaravan.GarageVehicle) { gv.GarageID = "" },
		"zero revision":         func(gv *opencaravan.GarageVehicle) { gv.RevisionVersion = 0 },
		"missing revision time": func(gv *opencaravan.GarageVehicle) { gv.RevisionTime = time.Time{} },
		"missing display name":  func(gv *opencaravan.GarageVehicle) { gv.DisplayName = "" },
		"zero capacity":         func(gv *opencaravan.GarageVehicle) { gv.Capacity = 0 },
		"missing signed_by":     func(gv *opencaravan.GarageVehicle) { gv.SignedBy = "" },
		"bad integrity":         func(gv *opencaravan.GarageVehicle) { gv.Integrity = &opencaravan.Integrity{} },
	}
	for name, mut := range cases {
		t.Run(name, func(t *testing.T) {
			gv := validGarageVehicle(t)
			mut(&gv)
			if err := gv.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestGarageVehicleCanonicalEncodingExcludesIntegrity(t *testing.T) {
	gv := validGarageVehicle(t)
	before, err := gv.CanonicalEncoding()
	if err != nil {
		t.Fatalf("encode before: %v", err)
	}
	gv.Integrity = &opencaravan.Integrity{
		Algorithm: "p256-ecdsa-sha256",
		KeyID:     "sha256:" + strings.Repeat("b", 64),
		Signature: base64.RawURLEncoding.EncodeToString([]byte("sig")),
	}
	after, err := gv.CanonicalEncoding()
	if err != nil {
		t.Fatalf("encode after: %v", err)
	}
	if string(before) != string(after) {
		t.Fatal("CanonicalEncoding must exclude Integrity")
	}
}
