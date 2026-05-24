package opencaravan

import "errors"

// Integrity describes the signature or message authentication data that makes
// a protocol object tamper-evident.
//
// Signature covers the canonical encoding of the carrying object excluding the
// Integrity field itself. Algorithm identifies the signature algorithm so a
// verifier can pick the correct primitive. KeyID identifies the issuer key a
// client can use to retrieve the verification public key.
//
// The same envelope is reused on every signed OpenCaravan object — journey
// invites today, server policy snapshots and future federation introductions
// next — so implementations can share a single signature-verification helper
// rather than re-implementing per object kind.
type Integrity struct {
	Algorithm string `json:"algorithm"`
	KeyID     string `json:"key_id,omitempty"`
	Signature string `json:"signature"`
}

// Validate reports whether integrity contains the fields needed to verify a
// signed object.
func (integrity Integrity) Validate() error {
	if integrity.Algorithm == "" {
		return errors.New("algorithm must be set")
	}
	if integrity.Signature == "" {
		return errors.New("signature must be set")
	}
	return nil
}
