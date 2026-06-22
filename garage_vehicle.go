package opencaravan

import (
	"errors"
	"fmt"
	"time"
)

// GarageVehicle is one vehicle entry in a [Garage]'s library. Persistent
// across journeys (unlike the journey-scoped [Vehicle]), and editable by
// any current accepted owner of its containing garage.
//
// GarageVehicle carries the vehicle's identity — what it is, what it looks
// like, who knows what about it — at the household level. The per-journey
// authorization model (who may drive it in a particular trip) lives on the
// journey-scoped [Vehicle] that participants upload when joining a journey.
// A client's "import this car into the journey" action copies relevant
// metadata from a GarageVehicle into a fresh Vehicle at upload time; there
// is no wire-level link between the two on a journey participant's view, so
// non-owner participants cannot correlate the same garage car appearing in
// multiple journeys. Cross-journey history aggregation for an owner's own
// view is downstream of this PR.
//
// Each GarageVehicle value is one revision in a monotonic chain. Revisions
// are signed by any current accepted owner of the containing garage; any
// owner may edit any vehicle in the garage. The server retains the full
// revision history.
type GarageVehicle struct {
	ID              UUID      `json:"id"`
	GarageID        UUID      `json:"garage_id"`
	RevisionVersion int       `json:"revision_version"`
	RevisionTime    time.Time `json:"revision_time"`
	DisplayName     string    `json:"display_name"`
	Make            string    `json:"make,omitempty"`
	Model           string    `json:"model,omitempty"`
	ModelYear       int       `json:"model_year,omitempty"`
	Color           string    `json:"color,omitempty"`
	Capacity        int       `json:"capacity"`
	// AvatarBlob references the compact / map-tile photo via
	// the protocol's content-addressed blob layer. See [BlobRef].
	AvatarBlob *BlobRef `json:"avatar_blob,omitempty"`
	// BannerBlob references the wide / detail-view photo via the
	// content-addressed blob layer.
	BannerBlob *BlobRef `json:"banner_blob,omitempty"`
	Notes      string   `json:"notes,omitempty"`
	// SignedBy names which current garage owner produced Integrity. The
	// server cross-checks against the named garage's owner list at
	// RevisionTime to confirm the signer was an accepted owner then.
	SignedBy  UUID       `json:"signed_by"`
	Integrity *Integrity `json:"integrity,omitempty"`
}

// Validate reports whether the garage vehicle has the required identity,
// garage reference, revision-bearing fields, capacity, image refs, and
// signed-envelope shape. Structural only — cryptographic verification and
// cross-checking against the garage's owner list at RevisionTime are
// consumer responsibilities.
func (gv GarageVehicle) Validate() error {
	if !gv.ID.Valid() {
		return errors.New("garage vehicle id must be a valid UUID")
	}
	if !gv.GarageID.Valid() {
		return errors.New("garage_id must be a valid UUID")
	}
	if gv.RevisionVersion < 1 {
		return errors.New("revision_version must be at least 1")
	}
	if gv.RevisionTime.IsZero() {
		return errors.New("revision_time must be set")
	}
	if gv.DisplayName == "" {
		return errors.New("display_name must be set")
	}
	if gv.Capacity < 1 {
		return errors.New("capacity must be at least 1")
	}
	if gv.AvatarBlob != nil {
		if err := gv.AvatarBlob.Validate(); err != nil {
			return fmt.Errorf("avatar_blob: %w", err)
		}
	}
	if gv.BannerBlob != nil {
		if err := gv.BannerBlob.Validate(); err != nil {
			return fmt.Errorf("banner_blob: %w", err)
		}
	}
	if !gv.SignedBy.Valid() {
		return errors.New("signed_by must be a valid UUID")
	}
	if gv.Integrity != nil {
		if err := gv.Integrity.Validate(); err != nil {
			return fmt.Errorf("integrity: %w", err)
		}
	}
	return nil
}

// CanonicalEncoding returns the deterministic byte sequence the signing
// owner signs to produce Integrity. The Integrity field itself is excluded.
func (gv GarageVehicle) CanonicalEncoding() ([]byte, error) {
	cp := gv
	cp.Integrity = nil
	return CanonicalJSON(cp)
}
