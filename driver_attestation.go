package opencaravan

import (
	"errors"
	"fmt"
	"time"
)

// DriverAttestation is the per-handoff signed payload a journey participant
// produces when taking over driving a vehicle at a waypoint. It is the unit
// of offline-tolerant authorization: a participant signs the attestation with
// their enrolled client cert, gossips it to other passengers over a peer
// transport, and any device that later reaches the server uploads the
// accumulated batch.
//
// The server replays each attestation: it verifies Integrity against the
// driver's enrolled cert, looks up the [VehicleACL] for ACLVersionConsulted,
// and checks that DriverUserID is in that ACL's AuthorizedDrivers list. An
// attestation against a since-revoked ACL is honored if the ACL was valid
// at the attestation's recorded version; a non-ACL driver is recorded with
// a downgraded trust flag when the vehicle's emergency rule permits it,
// rejected as an ACL violation otherwise. The server never deletes the
// recorded payload — chain of custody is information, not a gate.
//
// PriorAttestationHash optionally chains to the previous attestation the
// driver knew about (typically gossiped from the prior driver before they
// went offline). It is sha256:<lowercase hex of the SHA-256 of the prior
// attestation's CanonicalEncoding>. Two attestations sharing the same
// PriorAttestationHash signal a fork that the server flags and surfaces;
// when missing or unresolvable, ordering falls back to EffectiveTime.
type DriverAttestation struct {
	VehicleID            UUID      `json:"vehicle_id"`
	SegmentID            UUID      `json:"segment_id"`
	DriverUserID         UUID      `json:"driver_user_id"`
	EffectiveTime        time.Time `json:"effective_time"`
	ACLVersionConsulted  int       `json:"acl_version_consulted"`
	PriorAttestationHash *string   `json:"prior_attestation_hash,omitempty"`
	// Integrity is the driver's signature over CanonicalEncoding().
	Integrity *Integrity `json:"integrity,omitempty"`
}

// Validate reports whether the attestation has the required identity, segment
// linkage, ACL version, and structural envelope shape. Structural only —
// cryptographic verification and ACL-at-timestamp lookup are server / verifier
// responsibilities.
func (a DriverAttestation) Validate() error {
	if !a.VehicleID.Valid() {
		return errors.New("vehicle_id must be a valid UUID")
	}
	if !a.SegmentID.Valid() {
		return errors.New("segment_id must be a valid UUID")
	}
	if !a.DriverUserID.Valid() {
		return errors.New("driver_user_id must be a valid UUID")
	}
	if a.EffectiveTime.IsZero() {
		return errors.New("effective_time must be set")
	}
	if a.ACLVersionConsulted < 1 {
		return errors.New("acl_version_consulted must be at least 1")
	}
	if a.PriorAttestationHash != nil {
		if *a.PriorAttestationHash == "" {
			return errors.New("prior_attestation_hash must be non-empty when set")
		}
		if len(*a.PriorAttestationHash) < len("sha256:") || (*a.PriorAttestationHash)[:7] != "sha256:" {
			return errors.New("prior_attestation_hash must use the sha256: prefix")
		}
	}
	if a.Integrity != nil {
		if err := a.Integrity.Validate(); err != nil {
			return fmt.Errorf("integrity: %w", err)
		}
	}
	return nil
}

// CanonicalEncoding returns the deterministic byte sequence the driver signs
// to produce Integrity. The Integrity field itself is excluded.
//
// Note that PriorAttestationHash IS included in the signed input: chaining
// the prior attestation's hash into this attestation's signed bytes makes
// the chain tamper-evident at the signature layer.
func (a DriverAttestation) CanonicalEncoding() ([]byte, error) {
	cp := a
	cp.Integrity = nil
	return CanonicalJSON(cp)
}
