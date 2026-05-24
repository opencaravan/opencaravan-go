package opencaravan

import "time"

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
