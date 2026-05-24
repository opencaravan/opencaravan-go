package opencaravan

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ClientAppEnrollmentRequestType is the canonical type value for client app
// enrollment requests.
const ClientAppEnrollmentRequestType = "opencaravan.client_app_enrollment_request"

// ClientAppEnrollmentResponseType is the canonical type value for client app
// enrollment responses.
const ClientAppEnrollmentResponseType = "opencaravan.client_app_enrollment_response"

// ClientAppEnrollmentVersion is the current client app enrollment wire
// protocol version.
const ClientAppEnrollmentVersion = 1

// ClientAppEnrollment is the cryptographic credential record paired with a
// ClientApp. Each enrollment has the same ID as the ClientApp it credentials,
// so a ClientApp and a ClientAppEnrollment are two faces of one entity: the
// ClientApp describes who/what the installation is, and the
// ClientAppEnrollment carries the certificate chain that proves the
// installation controls a registered private key.
//
// The public key is intentionally omitted from this record. It is already
// inside CertificateChain[0] and derivable. Callers that need the public key
// outside an enrollment context use PublicKey directly.
type ClientAppEnrollment struct {
	ID     UUID `json:"id"`
	UserID UUID `json:"user_id"`
	// CertificateChain is the issued certificate chain, leaf first. The leaf
	// certificate's subject identifies the ClientApp and its public key is the
	// device's signing key.
	CertificateChain []string  `json:"certificate_chain"`
	IssuedTime       time.Time `json:"issued_time"`
	NotAfter         time.Time `json:"not_after"`
}

// Validate reports whether enrollment has the required identity fields, a
// PEM-shaped certificate chain, and consistent issuance and expiry times.
func (e ClientAppEnrollment) Validate() error {
	if !e.ID.Valid() {
		return errors.New("id must be a valid UUID")
	}
	if !e.UserID.Valid() {
		return errors.New("user_id must be a valid UUID")
	}
	if len(e.CertificateChain) == 0 {
		return errors.New("certificate_chain must contain at least the leaf certificate")
	}
	for i, pem := range e.CertificateChain {
		if err := validateCertificatePEM(pem); err != nil {
			return fmt.Errorf("certificate_chain[%d]: %w", i, err)
		}
	}
	if e.IssuedTime.IsZero() {
		return errors.New("issued_time must be set")
	}
	if e.NotAfter.IsZero() {
		return errors.New("not_after must be set")
	}
	if e.NotAfter.Before(e.IssuedTime) {
		return errors.New("not_after must not be before issued_time")
	}
	return nil
}

// KeyAttestationFormat names a recognized hardware-key attestation format.
//
// The protocol declares known formats so implementations agree on naming, but
// KeyAttestation.Validate does not reject unknown formats — the slot is
// forward-compatible with platform attestation schemes that arrive after this
// protocol version. Servers choose how strictly to verify each format.
type KeyAttestationFormat string

const (
	// KeyAttestationFormatAppleAppAttest is an Apple App Attest assertion.
	KeyAttestationFormatAppleAppAttest KeyAttestationFormat = "apple-app-attest"
	// KeyAttestationFormatAndroidKeyAttestation is an Android Keystore key
	// attestation certificate chain.
	KeyAttestationFormatAndroidKeyAttestation KeyAttestationFormat = "android-key-attestation"
	// KeyAttestationFormatWebAuthn is a WebAuthn attestation statement.
	KeyAttestationFormatWebAuthn KeyAttestationFormat = "webauthn"
)

// Valid reports whether the format is one of the OpenCaravan-known
// attestation formats. Unknown formats are valid on the wire (forward-compat),
// so callers that need a strict check use this helper explicitly.
func (f KeyAttestationFormat) Valid() bool {
	switch f {
	case KeyAttestationFormatAppleAppAttest, KeyAttestationFormatAndroidKeyAttestation, KeyAttestationFormatWebAuthn:
		return true
	default:
		return false
	}
}

// KeyAttestation carries an opaque platform-specific attestation that the
// client's signing key resides in hardware-backed storage.
//
// The protocol declares the slot; it does not verify the data. Servers may
// require an attestation of a specific format for high-trust scopes and
// invoke a format-specific verifier on the Data payload. Apps that lack
// hardware-backed storage omit KeyAttestation entirely.
type KeyAttestation struct {
	Format KeyAttestationFormat `json:"format"`
	// Data is base64-encoded format-specific payload. Encoding is base64
	// rather than base64url so existing attestation libraries on each platform
	// can ingest it without re-encoding.
	Data string `json:"data"`
}

// Validate reports whether attestation has a non-empty format and data
// payload. Unknown formats are accepted at the protocol layer.
func (a KeyAttestation) Validate() error {
	if strings.TrimSpace(string(a.Format)) == "" {
		return errors.New("format must be set")
	}
	if strings.TrimSpace(a.Data) == "" {
		return errors.New("data must be set")
	}
	return nil
}

// ClientAppEnrollmentRequest is the wire format an app sends to enroll a new
// ClientApp and obtain its credentialing ClientAppEnrollment.
//
// The app presents a one-shot InviteToken (server_registration scope), a
// CSR carrying its locally-generated public key, an optional human-readable
// display name, and an optional KeyAttestation. The server validates the
// invite, validates the CSR, signs the leaf certificate, and returns a
// ClientAppEnrollmentResponse.
type ClientAppEnrollmentRequest struct {
	Type        string `json:"type"`
	Version     int    `json:"version"`
	InviteToken string `json:"invite_token"`
	// CSRPEM is the PEM-encoded PKCS#10 certificate signing request. The
	// protocol requires PKCS#10 with a P-256 ECDSA public key, but parsing
	// and policy enforcement is implementation-level: this Validate only
	// checks that CSRPEM is a CERTIFICATE REQUEST PEM block.
	CSRPEM      string `json:"csr_pem"`
	DisplayName string `json:"display_name,omitempty"`
	// KeyAttestation is optional. When present, it proves the CSR's private
	// key resides in platform-managed hardware storage. Servers may make
	// attestation required for some scopes.
	KeyAttestation *KeyAttestation `json:"key_attestation,omitempty"`
}

// Validate reports whether the enrollment request has the required type,
// version, invite token, and PEM-shaped CSR. Optional attestation, if
// present, must also validate.
func (r ClientAppEnrollmentRequest) Validate() error {
	if r.Type != ClientAppEnrollmentRequestType {
		return fmt.Errorf("type must be %q", ClientAppEnrollmentRequestType)
	}
	if r.Version != ClientAppEnrollmentVersion {
		return fmt.Errorf("version must be %d", ClientAppEnrollmentVersion)
	}
	if strings.TrimSpace(r.InviteToken) == "" {
		return errors.New("invite_token must be set")
	}
	if err := validateCSRPEM(r.CSRPEM); err != nil {
		return fmt.Errorf("csr_pem: %w", err)
	}
	if r.KeyAttestation != nil {
		if err := r.KeyAttestation.Validate(); err != nil {
			return fmt.Errorf("key_attestation: %w", err)
		}
	}
	return nil
}

// ClientAppEnrollmentResponse is the server's reply to a successful
// ClientAppEnrollmentRequest. It carries the freshly-issued enrollment record
// and the server's CA chain so the app can pin the issuer for future
// validations.
type ClientAppEnrollmentResponse struct {
	Type    string `json:"type"`
	Version int    `json:"version"`
	// Enrollment is the issued certificate record. Enrollment.ID is also the
	// new ClientApp.ID; the app records both records under the same ID.
	Enrollment ClientAppEnrollment `json:"enrollment"`
	// ServerCAChain is the chain of issuer certificates above the leaf in
	// Enrollment.CertificateChain, root last. Apps pin the root for offline
	// peer-to-peer validation against other enrolled apps on the same server.
	ServerCAChain []string `json:"server_ca_chain"`
}

// Validate reports whether the enrollment response has the required type,
// version, enrollment record, and PEM-shaped CA chain.
func (r ClientAppEnrollmentResponse) Validate() error {
	if r.Type != ClientAppEnrollmentResponseType {
		return fmt.Errorf("type must be %q", ClientAppEnrollmentResponseType)
	}
	if r.Version != ClientAppEnrollmentVersion {
		return fmt.Errorf("version must be %d", ClientAppEnrollmentVersion)
	}
	if err := r.Enrollment.Validate(); err != nil {
		return fmt.Errorf("enrollment: %w", err)
	}
	if len(r.ServerCAChain) == 0 {
		return errors.New("server_ca_chain must contain at least the issuing CA certificate")
	}
	for i, pem := range r.ServerCAChain {
		if err := validateCertificatePEM(pem); err != nil {
			return fmt.Errorf("server_ca_chain[%d]: %w", i, err)
		}
	}
	return nil
}

// NewClientAppEnrollmentRequest returns a ClientAppEnrollmentRequest with the
// current type and version fields populated.
func NewClientAppEnrollmentRequest(inviteToken, csrPEM string) ClientAppEnrollmentRequest {
	return ClientAppEnrollmentRequest{
		Type:        ClientAppEnrollmentRequestType,
		Version:     ClientAppEnrollmentVersion,
		InviteToken: inviteToken,
		CSRPEM:      csrPEM,
	}
}

// NewClientAppEnrollmentResponse returns a ClientAppEnrollmentResponse with
// the current type and version fields populated.
func NewClientAppEnrollmentResponse(enrollment ClientAppEnrollment, serverCAChain []string) ClientAppEnrollmentResponse {
	return ClientAppEnrollmentResponse{
		Type:          ClientAppEnrollmentResponseType,
		Version:       ClientAppEnrollmentVersion,
		Enrollment:    enrollment,
		ServerCAChain: serverCAChain,
	}
}

func validateCertificatePEM(pem string) error {
	trimmed := strings.TrimSpace(pem)
	if trimmed == "" {
		return errors.New("must be set")
	}
	if !strings.Contains(trimmed, "-----BEGIN CERTIFICATE-----") || !strings.Contains(trimmed, "-----END CERTIFICATE-----") {
		return errors.New("must be a CERTIFICATE PEM block")
	}
	return nil
}

func validateCSRPEM(pem string) error {
	trimmed := strings.TrimSpace(pem)
	if trimmed == "" {
		return errors.New("must be set")
	}
	if !strings.Contains(trimmed, "-----BEGIN CERTIFICATE REQUEST-----") || !strings.Contains(trimmed, "-----END CERTIFICATE REQUEST-----") {
		return errors.New("must be a CERTIFICATE REQUEST PEM block")
	}
	return nil
}
