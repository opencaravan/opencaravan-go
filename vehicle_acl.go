package opencaravan

import (
	"errors"
	"fmt"
	"time"
)

// VehicleACL is the owner-signed payload published whenever the
// authorized-drivers list on a Vehicle changes. Each successive VehicleACL
// for a given vehicle carries a strictly higher ACLVersion, and a server
// or peer verifier retains every published VehicleACL so a DriverAttestation
// can validate against the ACL version that was current at its EffectiveTime
// rather than whatever ACL exists "now."
//
// Decoupling attestation validity from later ACL revisions is the load-bearing
// piece of the offline-tolerant model: a driver who attested at an earlier
// version is not retroactively unauthorized by a later revocation, and a
// driver who was authorized at attestation time stays authorized when their
// attestation finally syncs.
type VehicleACL struct {
	VehicleID         UUID                  `json:"vehicle_id"`
	OwnerUserID       UUID                  `json:"owner_user_id"`
	ACLVersion        int                   `json:"acl_version"`
	AuthorizedDrivers []UUID                `json:"authorized_drivers"`
	EmergencyRule     *VehicleEmergencyRule `json:"emergency_rule,omitempty"`
	EffectiveTime     time.Time             `json:"effective_time"`
	// Integrity is the owner's signature over CanonicalEncoding().
	Integrity *Integrity `json:"integrity,omitempty"`
}

// Validate reports whether the ACL update has the required identity,
// ownership, version-bearing fields, and structural envelope shape.
// Structural only — cryptographic verification is the consumer's
// responsibility once canonical bytes are reproduced via CanonicalEncoding.
func (a VehicleACL) Validate() error {
	if !a.VehicleID.Valid() {
		return errors.New("vehicle_id must be a valid UUID")
	}
	if !a.OwnerUserID.Valid() {
		return errors.New("owner_user_id must be a valid UUID")
	}
	if a.ACLVersion < 1 {
		return errors.New("acl_version must be at least 1")
	}
	for i, driver := range a.AuthorizedDrivers {
		if !driver.Valid() {
			return fmt.Errorf("authorized_drivers[%d] must be a valid UUID", i)
		}
	}
	if a.EmergencyRule != nil {
		if err := a.EmergencyRule.Validate(); err != nil {
			return fmt.Errorf("emergency_rule: %w", err)
		}
	}
	if a.EffectiveTime.IsZero() {
		return errors.New("effective_time must be set")
	}
	if a.Integrity != nil {
		if err := a.Integrity.Validate(); err != nil {
			return fmt.Errorf("integrity: %w", err)
		}
	}
	return nil
}

// CanonicalEncoding returns the deterministic byte sequence the owner signs
// to produce Integrity. The Integrity field itself is excluded.
func (a VehicleACL) CanonicalEncoding() ([]byte, error) {
	cp := a
	cp.Integrity = nil
	return CanonicalJSON(cp)
}
