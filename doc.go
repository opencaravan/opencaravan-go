// Package opencaravan defines draft Go types for the OpenCaravan protocol.
//
// OpenCaravan is an open protocol for coordinating group drives over networks.
// This module is intentionally protocol-focused: it contains shared vocabulary,
// wire-facing structs, and small validation helpers for journeys, segments,
// human participants, client apps, vehicles, shared media, and telemetry.
// Server storage, authentication persistence, and deployment concerns belong in
// implementing projects such as Spivot Server.
package opencaravan
