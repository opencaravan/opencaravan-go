package opencaravan

import (
	"errors"
	"strings"
	"time"
)

// MediaType describes the kind of participant-shared media attached to a
// journey or segment.
type MediaType string

const (
	// MediaPhoto means the shared media is a still image.
	MediaPhoto MediaType = "photo"
	// MediaVideo means the shared media is a video.
	MediaVideo MediaType = "video"
	// MediaAudio means the shared media is audio.
	MediaAudio MediaType = "audio"
	// MediaNote means the shared media is participant-authored text.
	MediaNote MediaType = "note"
	// MediaFile means the shared media is a generic file.
	MediaFile MediaType = "file"
)

// ImageResourceRef identifies an OpenCaravan image resource accepted by a
// server.
//
// The ID is enough for clients to derive the server's image fetch route. Digest
// gives clients a stable cache and integrity key without carrying generated
// URLs in profile, vehicle, or journey payloads.
type ImageResourceRef struct {
	ID           UUID   `json:"id"`
	Digest       string `json:"digest"`
	ContentType  string `json:"content_type"`
	WidthPixels  int    `json:"width_pixels,omitempty"`
	HeightPixels int    `json:"height_pixels,omitempty"`
}

// Validate reports whether ref contains a valid image identity and metadata.
func (ref ImageResourceRef) Validate() error {
	if !ref.ID.Valid() {
		return errors.New("image resource id must be a valid UUID")
	}
	if ref.Digest == "" {
		return errors.New("digest must be set")
	}
	contentType := strings.ToLower(ref.ContentType)
	if !strings.HasPrefix(contentType, "image/") || len(contentType) == len("image/") {
		return errors.New("content_type must be an image media type")
	}
	if ref.WidthPixels < 0 {
		return errors.New("width_pixels must not be negative")
	}
	if ref.HeightPixels < 0 {
		return errors.New("height_pixels must not be negative")
	}
	return nil
}

// SharedMedia describes media that a participant contributed to a journey.
//
// SegmentID is optional because some media belongs to the whole journey rather
// than one bounded segment. PolicyHash records the server policy document
// fingerprint that governed sharing when the media was accepted by the server.
type SharedMedia struct {
	ID                   UUID       `json:"id"`
	JourneyID            UUID       `json:"journey_id"`
	SegmentID            *UUID      `json:"segment_id,omitempty"`
	JourneyParticipantID UUID       `json:"journey_participant_id"`
	ClientAppID          UUID       `json:"client_app_id"`
	Type                 MediaType  `json:"type"`
	URL                  string     `json:"url"`
	ContentType          string     `json:"content_type,omitempty"`
	Caption              string     `json:"caption,omitempty"`
	PolicyHash           string     `json:"policy_hash"`
	CaptureTime          *time.Time `json:"capture_time,omitempty"`
	ShareTime            time.Time  `json:"share_time"`
}
