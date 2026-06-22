package opencaravan

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

// BlobRef is a content-addressed reference to an immutable blob of
// bytes hosted (and replicated) by an OpenCaravan server. The hash
// is both the identifier and the integrity check — clients that
// download the bytes can recompute sha256 and reject mismatches
// without needing a separate digest field.
//
// BlobRef supersedes the URL-based [ImageResourceRef] for fields
// where the bytes themselves are protocol-replicated rather than
// hosted on an external CDN. Used for vehicle avatar/banner photos
// today; the same shape extends to journey photo galleries and
// other shared-media use cases without protocol churn.
//
// Size and ContentType are carried alongside the hash so a peer
// receiving a bundle can render a placeholder (correct aspect
// ratio, "loading 2.3 MB photo…") before downloading and can
// reject obviously-wrong responses (claimed 5 MB jpeg returns 200
// KB of HTML).
type BlobRef struct {
	// Hash is the content hash in the canonical "sha256:<64-hex>"
	// format used everywhere else in the protocol.
	Hash string `json:"hash"`
	// Size is the byte length of the blob payload. Clients use this
	// for placeholders, range-fetch planning, and quota checks.
	Size int64 `json:"size"`
	// ContentType is the IANA media type the uploader supplied,
	// e.g., "image/jpeg" or "image/png". Lower-cased on the wire.
	ContentType string `json:"content_type"`
}

// Validate reports whether the ref has a structurally valid hash,
// non-negative size, and a canonical content type. Does not fetch
// the blob or verify the bytes match the hash — that's the
// consumer's responsibility after download.
//
// Because BlobRef sits inside the canonical bytes of [Vehicle] /
// [GarageVehicle] that get cryptographically signed, ContentType
// is enforced as strictly canonical: non-empty, lower-case, no
// leading or trailing whitespace. Two implementations that differ
// on whether they preserve "Image/JPEG" or normalize to
// "image/jpeg" would otherwise produce different signatures for
// the same blob, breaking peer verification silently.
func (b BlobRef) Validate() error {
	if err := ValidateCanonicalHash(b.Hash); err != nil {
		return fmt.Errorf("hash: %w", err)
	}
	if b.Size < 0 {
		return errors.New("size must not be negative")
	}
	if b.ContentType == "" {
		return errors.New("content_type must be set")
	}
	if b.ContentType != strings.ToLower(b.ContentType) {
		return errors.New("content_type must be lower-case")
	}
	if b.ContentType != strings.TrimSpace(b.ContentType) {
		return errors.New("content_type must not have leading or trailing whitespace")
	}
	return nil
}

// ValidateCanonicalHash reports whether s matches the protocol's
// canonical "sha256:<64-hex>" hash shape. The prefix is required
// (it leaves room for future algorithms without ambiguity) and the
// hex portion must be exactly 64 lower-case hex characters.
//
// Used by [BlobRef.Validate] and by other types that carry hash
// references (e.g., [DriverAttestation.PriorAttestationHash]).
func ValidateCanonicalHash(s string) error {
	const prefix = "sha256:"
	if !strings.HasPrefix(s, prefix) {
		return errors.New("must have sha256: prefix")
	}
	rest := s[len(prefix):]
	if len(rest) != 64 {
		return fmt.Errorf("hex portion must be 64 chars, got %d", len(rest))
	}
	if rest != strings.ToLower(rest) {
		return errors.New("hex portion must be lower-case")
	}
	if _, err := hex.DecodeString(rest); err != nil {
		return fmt.Errorf("hex decode: %w", err)
	}
	return nil
}
