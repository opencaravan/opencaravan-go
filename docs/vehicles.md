# Garages, Vehicles, Authorized Drivers, and Per-Segment Driver Attestation

This document is the canonical OpenCaravan specification for how vehicles are modeled, owned, signed, and operated — both in a user's persistent account-scoped library (the *garage* layer) and during a specific journey (the *journey* layer). It is implementation-language-agnostic. A competent engineer should be able to build a conformant client in Go, Swift, Kotlin, Rust, or TypeScript from this document alone, without reading any server's source code.

The Go reference types live in [`garage.go`](../garage.go), [`garage_vehicle.go`](../garage_vehicle.go), [`vehicle.go`](../vehicle.go), [`vehicle_acl.go`](../vehicle_acl.go), and [`driver_attestation.go`](../driver_attestation.go); the canonical-encoding helper lives in [`canonical.go`](../canonical.go). See [`protocol-model.md`](./protocol-model.md) for the broader protocol vocabulary (`UUID`, `ImageResourceRef`, `Integrity`, `User`, `Journey`).

## Two Layers

The protocol models vehicles in two distinct layers:

- **The garage layer** is account-scoped and persistent. A user enters their cars once, the entries live across many journeys, and a household can share authority over the whole library by listing multiple owners on a single `Garage`. This is where vehicle *identity* lives — what a car is, what it looks like, what notes the owners keep about it.
- **The journey layer** is journey-scoped and ephemeral. When a participant joins a journey they upload a fresh `Vehicle` for the trip. This is where *usage* lives — who can drive it during this trip, which segment it is currently in, which driver is currently behind the wheel. A journey `Vehicle` typically has its display name, photos, make/model, and capacity populated from a garage entry the participant has selected — but that linkage is a client-side concern, not exposed in the journey wire format. Other journey participants observe only the journey `Vehicle`; they cannot correlate it with any garage entry.

Keeping the layers separate preserves the privacy boundary between journeys (non-owner journey participants cannot see the same garage car appearing in multiple journeys) while enabling a household to maintain a shared "household garage" of vehicles. Wire-level linkage from journey Vehicle to garage entry is reserved for a later protocol version that explicitly opts in to cross-journey aggregation for owners.

## Design Intent

A `Vehicle` is a journey-scoped object that names a physical vehicle, the user who owns the record, the set of users authorized to drive it during the journey, and the cryptographic envelope that ties those facts to the owner's enrolled client certificate. A `DriverAttestation` is a signed, per-handoff record produced by a driver when they take over driving at a waypoint.

Three properties shape every other decision:

**Offline-tolerant by construction.** Mobile data is not available at every trailhead, summit, or parking lot. The protocol must function correctly with no server reachability at the moment of handoff. All authorization decisions are verifiable from state cached on the participating devices before they went offline.

**Server is eventual referee, not live gatekeeper.** When connectivity returns, the server replays cached attestations: it verifies signatures, looks up the ACL version that was current at each attestation's effective time, and records the result. Bad actors are detected after the fact; the trip is not gated on the server being able to authorize live.

**Secure protocol, ordinary UI.** The cryptographic machinery is rigorous because the protocol is the layer that resists adversarial behavior. Client UIs are encouraged to be friendly: a filtered list of "vehicles you can drive at this waypoint," surfaced from cached state. End users do not need to think about certificates, ACLs, or signatures — those facts are protocol-level, not UX-level.

## Garage Layer

### `Garage`

The account-scoped container that holds a household's library of `GarageVehicle` entries. A user with a single account may have a single garage; users in a household share a single garage by having multiple `Owners`. A user may participate in multiple garages — one household garage shared with a spouse, one project-car garage shared with weekend track buddies, and so on.

Each `Garage` value is one revision in a monotonic chain. Revisions are signed by any current accepted `Owner`, so any owner may add or remove other owners, rename the garage, or otherwise edit the container. The server retains the full revision history; the current state is the latest revision whose invited owners have all accepted.

| Field | JSON | Type | Required | Notes |
| --- | --- | --- | --- | --- |
| ID | `id` | UUID | yes | Client-generated, stable across revisions. |
| Name | `name` | string | yes | User-readable name. "Wheelsdown Household". |
| RevisionVersion | `revision_version` | int | yes | Monotonic; starts at 1, strictly increasing per ID. |
| RevisionTime | `revision_time` | RFC3339Nano UTC | yes | When this revision was signed. |
| Owners | `owners` | []GarageOwner | yes | At least one. Each entry names a user, when they were added, and when (if) they accepted the invitation. |
| SignedBy | `signed_by` | UUID | yes | Must reference an `Owner` whose `AcceptedTime` is set (a pending owner cannot sign updates). |
| Integrity | `integrity` | Integrity | yes (on wire) | Signature by the SignedBy owner over `CanonicalEncoding(Garage)`. |

### `GarageOwner`

| Field | JSON | Type | Required | Notes |
| --- | --- | --- | --- | --- |
| UserID | `user_id` | UUID | yes | The user this stake belongs to. |
| AddedTime | `added_time` | RFC3339Nano UTC | yes | Revision time when this owner was added. |
| AcceptedTime | `accepted_time` | RFC3339Nano UTC | no | Nil = pending acceptance (invitee has not yet published a matching `GarageOwnershipAcceptance`); set = active owner. When set, must not precede `AddedTime`. |

### `GarageOwnershipAcceptance`

The signed acknowledgement a newly-invited user publishes to accept a pending garage co-ownership invitation. The acceptance binds to a specific `Garage` revision (the revision in which the recipient was first added with `AcceptedTime` nil); replaying it against a different revision is rejected on a version mismatch.

| Field | JSON | Type | Required | Notes |
| --- | --- | --- | --- | --- |
| GarageID | `garage_id` | UUID | yes | Which garage. |
| RevisionVersionAccepted | `revision_version_accepted` | int | yes | Which revision invited me. |
| AccepterUserID | `accepter_user_id` | UUID | yes | The user accepting. Must match the cert that produced Integrity. |
| AcceptedTime | `accepted_time` | RFC3339Nano UTC | yes | When the recipient accepted. |
| Integrity | `integrity` | Integrity | yes (on wire) | Signature by the accepter's enrolled client cert. |

### `GarageVehicle`

One vehicle entry in a garage. Carries the persistent identity — display name, make/model/year/color, capacity, photos, owner notes. Distinct from the journey-scoped `Vehicle` below; see *Two Layers* for the relationship.

| Field | JSON | Type | Required | Notes |
| --- | --- | --- | --- | --- |
| ID | `id` | UUID | yes | Client-generated, stable across revisions. |
| GarageID | `garage_id` | UUID | yes | The owning garage. |
| RevisionVersion | `revision_version` | int | yes | Monotonic per ID. |
| RevisionTime | `revision_time` | RFC3339Nano UTC | yes | When this revision was signed. |
| DisplayName | `display_name` | string | yes | "Riley's Subaru", "The Blue Beast". |
| Make | `make` | string | no | Manufacturer. |
| Model | `model` | string | no | Model name. |
| ModelYear | `model_year` | int | no | Four-digit year. |
| Color | `color` | string | no | Free-form. |
| Capacity | `capacity` | int | yes | Total possible occupants including the driver. ≥ 1. |
| AvatarImage | `avatar_image` | ImageResourceRef | no | Square tile representation. |
| BannerImage | `banner_image` | ImageResourceRef | no | Wide header. |
| Notes | `notes` | string | no | Owner-visible free-form notes ("transmission rebuilt 2024"). |
| SignedBy | `signed_by` | UUID | yes | The garage owner who produced Integrity. Verifiers cross-check against the garage's owner list at `RevisionTime` to confirm the signer was an accepted owner then. |
| Integrity | `integrity` | Integrity | yes (on wire) | Signature by the SignedBy owner. |

### Garage authority model

The garage layer's authority rules:

1. **Any current accepted owner can sign updates.** Garage revisions, GarageVehicle revisions, ownership additions, ownership removals — all signed by any owner. The protocol does not distinguish "primary owner" or "admin"; all owners have equal authority.
2. **Adding a co-owner requires recipient consent.** An existing owner publishes a garage revision that adds the new user to `Owners` with `AcceptedTime` nil. The new user sees a pending invitation in their app; they accept by publishing a `GarageOwnershipAcceptance` for that revision. Server activates the revision (or surfaces it as active) once the acceptance is received. Without acceptance, the new user is not an owner — they cannot sign updates, they cannot see private garage contents.
3. **Removing a co-owner is unilateral.** Any current owner may sign a garage revision that excludes another owner. This intentionally permits a household to evict a lost or compromised account without the cooperation of that account. (Without unilateral removal, a lost-account problem becomes a permanent garage-ownership problem.)
4. **The sole remaining owner cannot voluntarily depart.** A revision that would leave `Owners` empty is rejected as a structurally invalid orphan. To dispose of a sole-owner garage, the owner deletes it (which is a separate, deferred operation).
5. **`SignedBy` must reference an accepted owner.** A pending invitee cannot sign garage updates — they would be signing on behalf of a household that has not yet ratified their participation.

### Lifecycle flows (garage layer)

#### G1. Garage creation

1. Client constructs a `Garage` with `RevisionVersion = 1`, the user as a single accepted owner (`AcceptedTime = AddedTime = now`), and signs.
2. Client uploads. Server verifies, persists, and returns the canonical garage.

#### G2. Adding a co-owner

1. Existing owner constructs `Garage` revision N+1 with the new user added to `Owners` (`AcceptedTime = nil`). Signs and uploads.
2. Server persists the revision as pending; new owner's app learns of the invitation.
3. New owner reviews and accepts: constructs a `GarageOwnershipAcceptance` referencing `(GarageID, RevisionVersionAccepted = N+1, AccepterUserID = self, AcceptedTime = now)`, signs, uploads.
4. Server verifies the acceptance, marks the revision active, sets the new owner's `AcceptedTime` to the acceptance time.

#### G3. Removing a co-owner

1. Any current accepted owner constructs a new `Garage` revision N+1 with the removed user absent from `Owners`. Signs and uploads.
2. Server verifies (signer is a current accepted owner; resulting `Owners` is non-empty; `RevisionVersion` strictly greater than the prior), persists.
3. Removed owner's view of the garage is revoked immediately. Future signatures by the removed user fail the accepted-owner check.

#### G4. Adding a vehicle to a garage

1. Any current accepted owner constructs a `GarageVehicle` with `RevisionVersion = 1`, populates metadata, signs with their cert, uploads.
2. Server verifies (signer is an accepted owner of the named garage), persists.

#### G5. Editing a vehicle

1. Any current accepted owner constructs the next monotonic `RevisionVersion` of the `GarageVehicle` with updated metadata, signs, uploads.
2. Server verifies and persists.

#### G6. Importing a garage vehicle into a journey

The "import" semantic the user-facing app surfaces:

1. App reads the user's accepted-owner garages and the `GarageVehicle` entries within them.
2. User picks one. App constructs a fresh journey `Vehicle` for the trip, copying `DisplayName`, `Make`, `Model`, `ModelYear`, `Color`, `Capacity`, `AvatarImage`, `BannerImage` from the garage entry.
3. User configures the journey-specific authorization (who may drive this vehicle in this trip, in `AuthorizedDrivers`), and signs as journey participant.
4. App uploads the journey `Vehicle` per the journey-layer flow in [Lifecycle Flows (journey layer)](#lifecycle-flows-journey-layer) below.

The journey `Vehicle` does not carry a wire-level reference back to the `GarageVehicle`. Other journey participants see only the journey vehicle.

## Journey Layer

### `Vehicle`

A signed metadata record for a vehicle participating in a specific journey. Typically populated from a `GarageVehicle` at upload time (see [G6](#g6-importing-a-garage-vehicle-into-a-journey)), but a participant may also enter a Vehicle fresh without any garage backing — the journey layer does not depend on the garage layer.

| Field | JSON | Type | Required | Notes |
| --- | --- | --- | --- | --- |
| ID | `id` | UUID | yes | Client-generated, server-maintained. Stable for the life of the journey. |
| DisplayName | `display_name` | string | yes | Operator-readable name. "The Blue Beast", "Riley's Subaru". |
| Make | `make` | string | no | Manufacturer. |
| Model | `model` | string | no | Model name. |
| ModelYear | `model_year` | int | no | Four-digit year. |
| Color | `color` | string | no | Free-form. |
| AvatarImage | `avatar_image` | ImageResourceRef | no | Square image rendered as the vehicle's tile representation. Clients should crop to circle. |
| BannerImage | `banner_image` | ImageResourceRef | no | Wide image used as a header on the vehicle's detail view. |
| OwnerUserID | `owner_user_id` | UUID | yes | The user whose enrolled client cert produced Integrity. Edit authority is scoped to this user — any client_app enrolled to this user may produce a fresh signed update. |
| Capacity | `capacity` | int | yes | Total possible occupants **including the driver**. A sedan is 5; a seven-seat minivan with six belts is 6. Capacity must be ≥ 1. |
| AuthorizedDrivers | `authorized_drivers` | []UUID | yes | User IDs authorized by the owner to drive this vehicle in this journey. May be empty (owner is sole driver). |
| ACLVersion | `acl_version` | int | yes | Monotonic counter starting at 1. Incremented whenever AuthorizedDrivers or EmergencyRule changes. DriverAttestations record the version they consulted. |
| EmergencyRule | `emergency_rule` | VehicleEmergencyRule | no | Owner-published fallback. Behavior is driven by `Kind`, not by presence: nil or `Kind = "none"` → no emergency policy, non-ACL attestations recorded as ACL violations; `Kind = "any_journey_participant"` → non-ACL attestations by journey participants recorded with a downgraded trust flag rather than rejected. Treating nil and `"none"` equivalently lets an owner publish an explicit "I considered the fallback question and chose no policy" signal distinct from "I haven't thought about it yet." |
| Integrity | `integrity` | Integrity | yes (on wire) | Signature by the owner's enrolled client cert over `CanonicalEncoding(Vehicle)`. Optional on a draft Vehicle that has not yet been signed; required for any server upload. |

### `VehicleEmergencyRule`

| Field | JSON | Type | Notes |
| --- | --- | --- | --- |
| Kind | `kind` | string enum | `"none"` (no emergency policy) or `"any_journey_participant"` (any participant in the journey may drive in an emergency, with downgraded trust flag). |

### `VehicleACL`

The owner-signed payload published when AuthorizedDrivers or EmergencyRule changes. Every published VehicleACL is retained by the server so DriverAttestations can validate against the ACL version that was current at their effective time.

| Field | JSON | Type | Required | Notes |
| --- | --- | --- | --- | --- |
| VehicleID | `vehicle_id` | UUID | yes | Identifies which Vehicle this ACL applies to. |
| OwnerUserID | `owner_user_id` | UUID | yes | Must match the Vehicle's OwnerUserID. |
| ACLVersion | `acl_version` | int | yes | Strictly greater than the previous published ACLVersion for this Vehicle. |
| AuthorizedDrivers | `authorized_drivers` | []UUID | yes | The new list. May be empty. |
| EmergencyRule | `emergency_rule` | VehicleEmergencyRule | no | The new emergency rule. |
| EffectiveTime | `effective_time` | RFC3339Nano UTC | yes | The instant the ACL takes effect. Attestations consult the ACL version current at *their* effective time, so an attestation pre-dating this EffectiveTime continues to validate against the prior version. |
| Integrity | `integrity` | Integrity | yes (on wire) | Owner signature. |

### `DriverAttestation`

A per-handoff signed payload a driver produces when taking over driving at a waypoint.

| Field | JSON | Type | Required | Notes |
| --- | --- | --- | --- | --- |
| VehicleID | `vehicle_id` | UUID | yes | Which Vehicle. |
| SegmentID | `segment_id` | UUID | yes | Which journey segment this attestation covers from now. |
| DriverUserID | `driver_user_id` | UUID | yes | The user_id taking the driver role. Must equal the user_id rolled up from the cert that produced Integrity. |
| EffectiveTime | `effective_time` | RFC3339Nano UTC | yes | When the driver takes over. Devices in the vehicle should use a shared time source; verifiers fall back to receive-order when timestamps are unreliable. |
| ACLVersionConsulted | `acl_version_consulted` | int | yes | The ACLVersion the driver's cached state validated against. Server replays validation against the same version, regardless of subsequent ACL revisions. |
| PriorAttestationHash | `prior_attestation_hash` | string | no | `"sha256:<lowercase 64 hex chars>"` — SHA-256 of the CanonicalEncoding of the prior attestation the driver knew about (typically gossiped from the previous driver before going offline). Optional because the first driver in a segment has no predecessor; conflict detection falls back to EffectiveTime ordering when absent. |
| Integrity | `integrity` | Integrity | yes (on wire) | Driver's signature. |

### `Integrity` (reused from the protocol vocabulary)

| Field | JSON | Type | Notes |
| --- | --- | --- | --- |
| Algorithm | `algorithm` | string | `"p256-ecdsa-sha256"` for v0 (matches client cert key type). |
| KeyID | `key_id` | string | `"sha256:<lowercase 64 hex chars>"` — SHA-256 fingerprint of the signing client cert. A verifier locates the cert from the journey's enrolled identity store. |
| Signature | `signature` | string | Base64url-encoded raw signature bytes (no padding). For P-256 ECDSA: ASN.1 DER encoding of `(r, s)`. |

## Signature Inputs

All signatures are computed over the `CanonicalEncoding(payload)` bytes — a deterministic JSON encoding the protocol defines explicitly so any conformant implementation reproduces identical bytes for the same input.

### Canonical encoding rules

1. **Start from a JSON representation of the payload with the Integrity field omitted.** A signature cannot cover itself; the Integrity field is excluded from its own signature input.
2. **All object keys are sorted lexicographically by Unicode code point at every level.**
3. **No insignificant whitespace.** No spaces or newlines between tokens. (Spaces *inside* string values are preserved verbatim.)
4. **Optional fields are omitted entirely when at their zero value.** A `null` is never emitted; the key simply is not present. This is the Go `,omitempty` convention; other languages must implement equivalent behavior.
5. **Time values are RFC3339Nano in UTC.** Example: `"2026-06-21T18:45:12.345678901Z"`. Trailing-zero precision is preserved as the language's standard library produces it; conformant implementations agree by going through `time.Time` (or equivalent) and serializing via the language's RFC3339Nano default.
6. **Integers and floats round-trip through IEEE 754 double-precision.** OpenCaravan field semantics keep integer values within the 2^53 mantissa range, so no precision is lost. Implementations that have a native integer JSON encoder may use it as long as the output matches Go's `encoding/json` for the same value.
7. **Arrays preserve their declared order.** Sorting applies only to object keys.

### Worked example

Given this Vehicle (before signing):

```json
{
  "id": "73ad7136-3592-4683-8bb9-dfb341b8e896",
  "display_name": "Riley's Subaru",
  "make": "Subaru",
  "owner_user_id": "2694b1c8-1d6e-4dda-b13a-eaf82fc5a31a",
  "capacity": 5,
  "authorized_drivers": ["86a9e7f7-081d-4d7a-9aa1-2fa6034abb70"],
  "acl_version": 1
}
```

The canonical encoding is:

```
{"acl_version":1,"authorized_drivers":["86a9e7f7-081d-4d7a-9aa1-2fa6034abb70"],"capacity":5,"display_name":"Riley's Subaru","id":"73ad7136-3592-4683-8bb9-dfb341b8e896","make":"Subaru","owner_user_id":"2694b1c8-1d6e-4dda-b13a-eaf82fc5a31a"}
```

The owner signs SHA-256 of those exact bytes with their client cert's private key, then attaches the resulting Integrity field. Verifiers reproduce the canonical bytes from the received Vehicle (with Integrity stripped), recompute SHA-256, and verify against the cert identified by Integrity.KeyID.

A conformant test that exercises this round-trip in another language: produce the example bytes above, sign with a fresh P-256 keypair, verify, then mutate one byte and confirm verification fails. The Go reference test is `TestVehicleSignVerifyRoundTrip` in [`vehicle_auth_test.go`](../vehicle_auth_test.go).

## Lifecycle Flows (journey layer)

### 1. Vehicle upload

A user joins a journey and uploads a Vehicle for the duration:

1. Client constructs the Vehicle struct populated with the journey-scoped UUID, owner identity, capacity, and the AuthorizedDrivers list (typically just the owner at first; can be expanded later).
2. Client computes `CanonicalEncoding(vehicle)`, signs the result with its enrolled client cert's private key, and attaches Integrity.
3. Client POSTs the Vehicle to the server's vehicle-upload endpoint.
4. Server verifies Integrity against the OwnerUserID's enrolled client cert (identified by Integrity.KeyID), persists the Vehicle, retains the initial ACL state as ACLVersion 1.

### 2. ACL update

The owner adds or removes authorized drivers, or publishes/changes the emergency rule:

1. Client constructs a VehicleACL with the next monotonic ACLVersion, the new AuthorizedDrivers list, the EffectiveTime, and the EmergencyRule.
2. Client signs CanonicalEncoding(acl), attaches Integrity.
3. Client uploads the VehicleACL.
4. Server verifies, persists the new VehicleACL, retains the prior versions. Subsequent DriverAttestations may reference any retained version.

### 3. Offline driver handoff

At a waypoint with no server reachability:

1. The new driver (a user in the cached AuthorizedDrivers list for the current ACLVersion) constructs a DriverAttestation: VehicleID, SegmentID, their own user_id as DriverUserID, EffectiveTime = now (wall clock), ACLVersionConsulted = the cached ACL version, PriorAttestationHash = SHA-256 of the canonical encoding of the prior attestation if known.
2. The new driver signs CanonicalEncoding(attestation) with their client cert's private key, attaches Integrity.
3. The new driver broadcasts the attestation to other devices in the vehicle over an in-vehicle peer transport (BLE / Multipeer Connectivity / Wi-Fi Direct).
4. Each receiving device locally verifies: signature chains to a known journey-participant cert, signer's user_id is in the cached AuthorizedDrivers for ACLVersionConsulted, EffectiveTime is monotonic relative to prior cached attestations.
5. Each device retains the attestation in its local store; any single device that later reaches the server uploads on behalf of the group.

### 4. Sync

Whichever device first reaches mobile data uploads its accumulated batch of attestations:

1. POST the batch to the per-vehicle attestation endpoint.
2. Server iterates: verify Integrity, look up the VehicleACL for ACLVersionConsulted, check DriverUserID ∈ that ACL's AuthorizedDrivers.
3. For each, the server records a verification status:
   - `verified` — signature OK, ACL membership confirmed
   - `acl_violation` — signature OK, but the driver is not in the consulted ACL version and the Vehicle's EmergencyRule does not permit fallback
   - `emergency_authorized` — signature OK, driver not in ACL, but EmergencyRule = `any_journey_participant` and the signer is a journey participant
   - `signature_invalid` — Integrity does not validate against the named cert
4. Attestations are idempotent on (signature, vehicle_id). Duplicate uploads from gossiping clients are absorbed.

### 5. Server replay (any time after sync)

A server, a third-party observer, or a future audit tool reconstructs the driver timeline for a vehicle:

1. List all DriverAttestations for the vehicle.
2. Order by EffectiveTime; resolve hash chains where present.
3. For each interval (this attestation's EffectiveTime → next attestation's EffectiveTime), the recorded driver is DriverUserID.
4. Telemetry batches recorded during each interval carry the attestation hash they were submitted under (see telemetry batch optional field `driver_attestation_hash`); chain of custody is preserved.

## Server Semantics

### Validation rules

- **Owner-signed Vehicle uploads** must validate Integrity against the OwnerUserID's enrolled client cert at upload time. The cert may be revoked later; the upload at the time was valid.
- **VehicleACL updates** must (a) reference an existing Vehicle, (b) have ACLVersion strictly greater than any prior published ACLVersion for that vehicle, (c) be signed by the Vehicle's OwnerUserID.
- **DriverAttestations** validate per the verification-status logic in the sync flow above. No attestation is rejected outright — every received attestation is recorded so the audit trail survives. The verification status is metadata, not a gate.

### Conflict detection

Two attestations sharing the same `prior_attestation_hash` indicate a fork — two drivers both claimed the driver role downstream of the same predecessor. The server flags the affected interval with a `driver_conflict` marker and surfaces both attestations via the API. The protocol does not arbitrate which attestation was "really" the driver; that is a downstream resolution layer's concern (typically the journey host or a human-resolved UI).

When `prior_attestation_hash` is absent on conflicting attestations, the server falls back to ordering by `effective_time`; ties are surfaced as conflicts.

### Ownership freeze

When the OwnerUserID ceases to be a participant in the journey (left, removed, or had their journey participation revoked), subsequent edits to the Vehicle's metadata and ACL are rejected. The Vehicle remains in the journey for its remaining lifetime, but ACL changes are no longer accepted from any party. This preserves authorship without requiring the owner to babysit; it is the simplest answer to "what happens when the owner departs."

### Flag-not-block policy

The protocol's recording bias is to retain every signed payload. The cryptography establishes provenance and the recorded verification status describes the trust level of each record. Refusing to record a payload would damage the audit trail; downgrading its trust label and surfacing the downgrade via the API preserves the data while telling consumers honestly what is and is not load-bearing.

## Client Semantics

### Filtered-list construction

For a "vehicles you can drive at this waypoint" picker, the client:

1. Holds the cached set of journey Vehicles (downloaded at journey start, refreshed when ACL updates sync in).
2. Filters by: AuthorizedDrivers contains this user_id at the current cached ACLVersion, OR Vehicle.EmergencyRule.Kind = `any_journey_participant`.
3. Presents the filtered list to the user. The user picks one; the client constructs and signs a DriverAttestation under the hood and broadcasts it.

The client does not surface ACL membership status, certificate fingerprints, signature validity, or any other cryptographic detail to the end user. The UX is "tap to drive."

### Gossip-and-sync

When the device returns to connectivity, it uploads its local batch of attestations. Any device's upload is sufficient; conflict-free idempotency on the server absorbs duplicate uploads from devices that all observed the same attestations.

### UX recommendations (non-binding)

These are recommendations rather than protocol requirements; conformant clients may render this state however suits their application:

- Surface a small "synced" or "pending sync" indicator on the driver swap UI so the user understands whether their attestation has reached the server.
- For attestations recorded with `acl_violation` or `emergency_authorized` status, surface the trust label to the journey host post-trip rather than to the driver mid-trip. The driver is on the road; the host can handle audit concerns at next stop.
- For `driver_conflict` markers, surface both attestations to the journey host and prompt for human resolution. Do not auto-pick.

## Failure Modes

### ACL revocation between attestation and sync

A driver signs an attestation at time T₁ with ACLVersionConsulted = N. Between T₁ and the eventual sync, the owner publishes ACLVersion N+1 that removes this driver from AuthorizedDrivers. At sync, the server consults ACLVersion N (not N+1) per the EffectiveTime → ACL-version lookup, and records the attestation as `verified`. The driver was authorized at the time; later revocation does not retroactively invalidate.

### Forked attestations

Two passengers offline both claim the driver role with the same `prior_attestation_hash`. Both attestations are valid per signature and ACL membership. The server records both as `verified` and flags the affected segment with `driver_conflict`. Resolution is delegated to a downstream layer; the protocol's job is honest reporting.

### Clock skew between offline devices

Devices offline for extended periods may drift apart in wall-clock time. Attestation EffectiveTime is the signer's wall clock at attestation. When conflict detection falls back to EffectiveTime ordering (no PriorAttestationHash chain), order is "what each device reported." The server may use server-side `received_at` as a tiebreaker but the protocol does not require this — implementations are free to choose tiebreaker semantics in their own audit-trail UI.

### Unauthorized claim (non-ACL participant)

A signer not in the consulted ACL produces an attestation. If the Vehicle has no EmergencyRule, status = `acl_violation`. If EmergencyRule.Kind = `any_journey_participant` and the signer is a journey participant, status = `emergency_authorized`. Either way, the attestation is recorded; telemetry submitted under it is retained with the same trust label.

### Ownership freeze

After the OwnerUserID departs the journey, further VehicleACL updates from any party are rejected. DriverAttestations continue to be accepted against the last-published ACL; the vehicle can still be driven, the driver list just cannot evolve.

### Attestation against a future ACL version

A driver attempts to claim ACLVersionConsulted = N+1 when no such version has been published. Status = `acl_violation`. (The driver could not have observed an ACL that did not exist.)

## Extension Points (Reserved for Future Protocol Versions)

The following are deliberately out of scope for v0.1.x; the wire format reserves space and discipline for them but does not implement them:

- **Co-signed handoff mode.** A stricter opt-in mode where the new driver's attestation only validates if the prior driver counter-signs. Eliminates forks but breaks if the prior driver is unable/unwilling. Future Vehicle field `handoff_policy` will name the modes.
- **P2P gossip transport.** The protocol shape supports it (self-contained signed payloads, no server mutation on accept); the BLE / Multipeer / Wi-Fi Direct transport itself is a downstream concern.
- **Wire-level garage→journey linkage.** v0.1.x keeps the layers independent at the wire level: a journey `Vehicle` is its own thing, populated client-side from a `GarageVehicle` at upload time. A future version may add an opt-in `garage_vehicle_id` field on `Vehicle` for owners who want cross-journey aggregation in their own dashboards. Non-owner participants would not observe the linkage.
- **Garage update sync to past journeys.** Editing a `GarageVehicle` (new photo, capacity correction) does not retroactively change journey `Vehicle`s that were uploaded under the prior garage revision. A future version may surface "this car was used in N journeys with these date ranges" in an owner's dashboard view, computed server-side from the linkage field above.
- **Garage deletion.** v0.1.x has no formal delete payload for a `Garage` or `GarageVehicle`. Server implementations may permit administrative deletion by sole-owner accounts; a protocol-level signed deletion payload is reserved for a future version.
- **Multi-server garage federation.** A garage exists on a single server. Cross-server sharing (federating a household garage across servers an owner is enrolled with) is a federation concern, deferred.
- **Attestation revocation.** A driver who wants to formally withdraw an earlier attestation (e.g., "I claimed to drive but then I didn't") can today rely on a subsequent attestation overriding the prior. A formal revocation payload may be added when the use case sharpens.
- **Multi-occupant signed attestations.** A v2 may permit multiple occupants to co-sign per-segment attestations recording the full vehicle roster, not just the driver. The current SegmentVehicle + VehicleOccupant types already accommodate the data shape; the cryptographic envelope is the missing piece.

## Versioning

This document corresponds to OpenCaravan protocol version `0.1.x` (additive; vehicles + driver attestations are an extension, not a breaking change to existing types). The protocol-version constant on the server's `/v1/server` endpoint follows the OpenCaravan protocol-versus-implementation decoupling discipline; this spec does not require a bump until a wire-breaking change to these types lands.
