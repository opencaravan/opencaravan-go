package opencaravan

import (
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
)

const nilUUID = "00000000-0000-0000-0000-000000000000"

// UUID identifies OpenCaravan protocol objects with a canonical UUID string.
//
// UUID values marshal as lowercase RFC 4122-style text. The nil UUID is not a
// valid OpenCaravan identifier because protocol IDs are expected to name real
// journey, participant, client, vehicle, segment, media, or telemetry objects.
type UUID string

// NewUUID returns a random version 4 UUID suitable for assigning to an
// OpenCaravan protocol object.
func NewUUID() (UUID, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("read random UUID bytes: %w", err)
	}

	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	return UUID(fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])), nil
}

// ParseUUID parses and canonicalizes an OpenCaravan UUID.
func ParseUUID(s string) (UUID, error) {
	id := UUID(strings.ToLower(s))
	if !id.Valid() {
		return "", fmt.Errorf("invalid UUID %q", s)
	}
	return id, nil
}

// Valid reports whether id is a non-nil UUID-shaped value.
func (id UUID) Valid() bool {
	s := string(id)
	if len(s) != 36 || s == nilUUID {
		return false
	}

	for i, r := range s {
		switch i {
		case 8, 13, 18, 23:
			if r != '-' {
				return false
			}
		default:
			if !isHex(r) {
				return false
			}
		}
	}

	return true
}

// MarshalText returns the canonical text form of id.
func (id UUID) MarshalText() ([]byte, error) {
	parsed, err := ParseUUID(string(id))
	if err != nil {
		return nil, err
	}
	return []byte(parsed), nil
}

// UnmarshalText parses a canonical UUID text value into id.
func (id *UUID) UnmarshalText(text []byte) error {
	if id == nil {
		return errors.New("unmarshal UUID into nil pointer")
	}

	parsed, err := ParseUUID(string(text))
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

func isHex(r rune) bool {
	return r >= '0' && r <= '9' || r >= 'a' && r <= 'f' || r >= 'A' && r <= 'F'
}
