package opencaravan

import (
	"errors"
	"fmt"
	"strings"
)

// KeyAlgorithm names a public-key algorithm recognized by OpenCaravan
// implementations.
//
// Initial OpenCaravan implementations target P-256 ECDSA because it is the
// only key type with hardware-backed storage on iOS Secure Enclave, Android
// Keystore (universally), and WebAuthn. New algorithms may be added in
// future protocol versions; the typed constant guards against typos while
// leaving the door open for additions.
type KeyAlgorithm string

const (
	// KeyAlgorithmP256ECDSA is NIST P-256 (secp256r1) with ECDSA signatures.
	KeyAlgorithmP256ECDSA KeyAlgorithm = "p256-ecdsa"
)

// Valid reports whether the key algorithm is a known OpenCaravan value.
func (a KeyAlgorithm) Valid() bool {
	switch a {
	case KeyAlgorithmP256ECDSA:
		return true
	default:
		return false
	}
}

// PublicKey is a protocol-level envelope for a public key.
//
// PublicKey carries a PEM-encoded SubjectPublicKeyInfo block and an explicit
// algorithm tag. The algorithm tag lets verifiers select the right primitive
// without parsing the key, and lets implementations reject unknown algorithms
// at protocol boundaries before reaching crypto code.
//
// PublicKey is used in protocol contexts where a key needs to travel
// independently of any certificate — for example, advertising a server signing
// key, or carrying a key in flight before a CSR has been signed. Once a
// certificate exists, the key is derivable from the certificate and the
// separate PublicKey envelope is not required.
type PublicKey struct {
	Algorithm KeyAlgorithm `json:"algorithm"`
	// KeyPEM is the PEM-encoded SubjectPublicKeyInfo. For P-256 ECDSA the
	// PEM block label is "PUBLIC KEY".
	KeyPEM string `json:"key_pem"`
}

// Validate reports whether the public key has a known algorithm and a
// PEM-shaped key block. Validate performs only format-level checks — it does
// not parse the PEM contents or verify that the key matches the declared
// algorithm; implementations enforce the deeper checks at their crypto
// boundary.
func (k PublicKey) Validate() error {
	if !k.Algorithm.Valid() {
		return errors.New("algorithm must be a known OpenCaravan value")
	}
	if strings.TrimSpace(k.KeyPEM) == "" {
		return errors.New("key_pem must be set")
	}
	if !strings.Contains(k.KeyPEM, "-----BEGIN PUBLIC KEY-----") {
		return fmt.Errorf("key_pem must contain a PUBLIC KEY PEM block")
	}
	if !strings.Contains(k.KeyPEM, "-----END PUBLIC KEY-----") {
		return fmt.Errorf("key_pem must contain a complete PUBLIC KEY PEM block")
	}
	return nil
}
