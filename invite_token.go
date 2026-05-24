package opencaravan

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"
)

// InviteTokenBytes is the number of cryptographically random bytes in a
// generated invite token value.
const InviteTokenBytes = 32

// InviteToken is the server-issued secret capability carried by an invite of
// any scope. Both journey invites and server-registration invites share this
// shape so implementations can build a single token issuance and redemption
// path.
type InviteToken struct {
	Value          string    `json:"value"`
	ExpirationTime time.Time `json:"expiration_time"`
}

// NewInviteToken returns a cryptographically random invite token.
//
// The token value contains 256 bits of randomness encoded as unpadded base64url
// text so it can travel safely in URLs, QR codes, JSON, and platform share
// payloads.
func NewInviteToken(expirationTime time.Time) (InviteToken, error) {
	if expirationTime.IsZero() {
		return InviteToken{}, errors.New("invite token expiration_time must be set")
	}

	b := make([]byte, InviteTokenBytes)
	if _, err := rand.Read(b); err != nil {
		return InviteToken{}, fmt.Errorf("read random invite token bytes: %w", err)
	}

	return InviteToken{
		Value:          base64.RawURLEncoding.EncodeToString(b),
		ExpirationTime: expirationTime,
	}, nil
}

// Validate reports whether token contains a secret value and expiration.
func (token InviteToken) Validate() error {
	if token.Value == "" {
		return errors.New("invite token value must be set")
	}
	tokenBytes, err := base64.RawURLEncoding.DecodeString(token.Value)
	if err != nil {
		return fmt.Errorf("invite token value must be unpadded base64url: %w", err)
	}
	if len(tokenBytes) != InviteTokenBytes {
		return fmt.Errorf("invite token must contain %d random bytes", InviteTokenBytes)
	}
	if token.ExpirationTime.IsZero() {
		return errors.New("invite token expiration_time must be set")
	}
	return nil
}
