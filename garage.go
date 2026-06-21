package opencaravan

import (
	"errors"
	"fmt"
	"time"
)

// Garage is the account-scoped, persistent, multi-owner container that holds
// a user's library of [GarageVehicle] entries. A user with a single account
// usually has a single garage; users in a household share a single garage by
// having multiple Owners. A user may participate in multiple garages — one
// household garage shared with their spouse, one project-car garage shared
// with weekend track buddies, and so on.
//
// Each Garage value is one revision in a monotonic chain. Revisions are
// signed by any current Owner, so any owner may add or remove other owners,
// rename the garage, or otherwise edit the container. The server retains the
// full revision history; the current state is the latest revision whose
// invited Owners have all accepted (pending acceptances are visible to the
// invitee but the revision is not yet active).
//
// Ownership additions require recipient consent via [GarageOwnershipAcceptance]
// — a co-owner cannot be added unilaterally, which prevents an attacker
// from spamming unwanted entries into another user's garage. Ownership
// removals are unilateral: any current owner may remove any other owner.
// This matches the household trust model; the protocol's answer to a lost
// or compromised account is "an existing co-owner removes it from the
// garage" rather than "no one can ever remove this account."
type Garage struct {
	ID              UUID          `json:"id"`
	Name            string        `json:"name"`
	RevisionVersion int           `json:"revision_version"`
	RevisionTime    time.Time     `json:"revision_time"`
	Owners          []GarageOwner `json:"owners"`
	SignedBy        UUID          `json:"signed_by"`
	Integrity       *Integrity    `json:"integrity,omitempty"`
}

// GarageOwner names one user's ownership stake in a Garage. AcceptedTime is
// nil while the invitation is pending; non-nil once the recipient has
// published a matching [GarageOwnershipAcceptance].
//
// A garage revision in which a new owner is added with AcceptedTime nil
// represents a pending invitation: the recipient sees it in their app and
// may accept or decline; non-recipients do not see the garage at all.
// Once the recipient accepts, a subsequent revision updates AcceptedTime
// (or the server materializes the accepted state without a new revision —
// see docs/vehicles.md for the wire-level options).
type GarageOwner struct {
	UserID       UUID       `json:"user_id"`
	AddedTime    time.Time  `json:"added_time"`
	AcceptedTime *time.Time `json:"accepted_time,omitempty"`
}

// GarageOwnershipAcceptance is the signed acknowledgement a newly-invited
// user publishes to accept a pending garage co-ownership invitation. The
// acceptance binds to a specific Garage revision (the revision in which
// the recipient was first added with AcceptedTime nil); replaying it
// against a different revision is rejected on a version mismatch.
type GarageOwnershipAcceptance struct {
	GarageID                UUID       `json:"garage_id"`
	RevisionVersionAccepted int        `json:"revision_version_accepted"`
	AccepterUserID          UUID       `json:"accepter_user_id"`
	AcceptedTime            time.Time  `json:"accepted_time"`
	Integrity               *Integrity `json:"integrity,omitempty"`
}

// Validate reports whether the garage has the required identity, ownership
// list, revision-bearing fields, and signed-envelope shape. Structural
// only — cryptographic verification is the consumer's responsibility once
// canonical bytes are reproduced via [Garage.CanonicalEncoding].
//
// At least one owner must be present (a garage with zero owners is
// orphaned). The SignedBy user must appear in Owners as an accepted owner;
// pending owners cannot sign updates.
func (g Garage) Validate() error {
	if !g.ID.Valid() {
		return errors.New("garage id must be a valid UUID")
	}
	if g.Name == "" {
		return errors.New("garage name must be set")
	}
	if g.RevisionVersion < 1 {
		return errors.New("revision_version must be at least 1")
	}
	if g.RevisionTime.IsZero() {
		return errors.New("revision_time must be set")
	}
	if len(g.Owners) == 0 {
		return errors.New("garage must have at least one owner")
	}
	seen := make(map[UUID]struct{}, len(g.Owners))
	signerAccepted := false
	for i, owner := range g.Owners {
		if err := owner.Validate(); err != nil {
			return fmt.Errorf("owners[%d]: %w", i, err)
		}
		if _, dup := seen[owner.UserID]; dup {
			return fmt.Errorf("owners[%d]: duplicate user_id %q", i, owner.UserID)
		}
		seen[owner.UserID] = struct{}{}
		if owner.UserID == g.SignedBy && owner.AcceptedTime != nil {
			signerAccepted = true
		}
	}
	if !g.SignedBy.Valid() {
		return errors.New("signed_by must be a valid UUID")
	}
	if !signerAccepted {
		return errors.New("signed_by must reference an accepted owner")
	}
	if g.Integrity != nil {
		if err := g.Integrity.Validate(); err != nil {
			return fmt.Errorf("integrity: %w", err)
		}
	}
	return nil
}

// CanonicalEncoding returns the deterministic byte sequence the signing
// owner signs to produce Integrity. The Integrity field itself is excluded.
func (g Garage) CanonicalEncoding() ([]byte, error) {
	cp := g
	cp.Integrity = nil
	return CanonicalJSON(cp)
}

// Validate reports whether the owner stake names a user, an addition time,
// and a structurally-valid acceptance time when set.
func (o GarageOwner) Validate() error {
	if !o.UserID.Valid() {
		return errors.New("user_id must be a valid UUID")
	}
	if o.AddedTime.IsZero() {
		return errors.New("added_time must be set")
	}
	if o.AcceptedTime != nil && o.AcceptedTime.IsZero() {
		return errors.New("accepted_time must be a non-zero time when set")
	}
	if o.AcceptedTime != nil && o.AcceptedTime.Before(o.AddedTime) {
		return errors.New("accepted_time must not be before added_time")
	}
	return nil
}

// Validate reports whether the acceptance carries the required garage
// reference, accepter identity, and signed-envelope shape.
func (a GarageOwnershipAcceptance) Validate() error {
	if !a.GarageID.Valid() {
		return errors.New("garage_id must be a valid UUID")
	}
	if a.RevisionVersionAccepted < 1 {
		return errors.New("revision_version_accepted must be at least 1")
	}
	if !a.AccepterUserID.Valid() {
		return errors.New("accepter_user_id must be a valid UUID")
	}
	if a.AcceptedTime.IsZero() {
		return errors.New("accepted_time must be set")
	}
	if a.Integrity != nil {
		if err := a.Integrity.Validate(); err != nil {
			return fmt.Errorf("integrity: %w", err)
		}
	}
	return nil
}

// CanonicalEncoding returns the deterministic byte sequence the accepter
// signs to produce Integrity. The Integrity field itself is excluded.
func (a GarageOwnershipAcceptance) CanonicalEncoding() ([]byte, error) {
	cp := a
	cp.Integrity = nil
	return CanonicalJSON(cp)
}
