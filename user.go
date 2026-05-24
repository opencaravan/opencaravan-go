package opencaravan

import (
	"errors"
	"fmt"
	"net/url"
)

// User describes a server-scoped OpenCaravan identity.
//
// Servers assign user IDs per connection or registration. A person may have
// different user IDs and different profile details on different servers.
type User struct {
	ID      UUID        `json:"id"`
	Profile UserProfile `json:"profile"`
	// DeletionAfterInactivitySeconds is an optional duration after which the
	// server may delete this user record if no activity resets the timer.
	DeletionAfterInactivitySeconds *int64      `json:"deletion_after_inactivity_seconds,omitempty"`
	ClientApps                     []ClientApp `json:"client_apps,omitempty"`
}

// UserProfile describes client-supplied profile information a server may
// republish to authorized journey participants.
//
// Contacts are explicitly profile-visible. Private account recovery fields,
// billing data, and server-only authentication metadata do not belong here.
// Clients may mirror one profile across servers or tailor profile details for
// each server registration.
type UserProfile struct {
	DisplayName string               `json:"display_name"`
	AvatarURL   string               `json:"avatar_url,omitempty"`
	BannerURL   string               `json:"banner_url,omitempty"`
	Bio         string               `json:"bio,omitempty"`
	AccentColor string               `json:"accent_color,omitempty"`
	Links       []UserProfileLink    `json:"links,omitempty"`
	Contacts    []UserProfileContact `json:"contacts,omitempty"`
}

// UserProfileLink describes one user-supplied public profile link.
type UserProfileLink struct {
	Kind  string `json:"kind,omitempty"`
	Label string `json:"label,omitempty"`
	URL   string `json:"url"`
}

// UserProfileContactKind describes the type of profile-visible contact method.
type UserProfileContactKind string

const (
	// UserProfileContactPhone means the contact URI starts a phone call.
	UserProfileContactPhone UserProfileContactKind = "phone"
	// UserProfileContactSMS means the contact URI starts an SMS conversation.
	UserProfileContactSMS UserProfileContactKind = "sms"
	// UserProfileContactEmail means the contact URI starts an email message.
	UserProfileContactEmail UserProfileContactKind = "email"
	// UserProfileContactURL means the contact URI opens a general URL.
	UserProfileContactURL UserProfileContactKind = "url"
	// UserProfileContactOther means the contact URI is a server- or
	// client-understood contact method outside the standard set.
	UserProfileContactOther UserProfileContactKind = "other"
)

// Valid reports whether kind is a known OpenCaravan profile contact kind.
func (kind UserProfileContactKind) Valid() bool {
	switch kind {
	case UserProfileContactPhone, UserProfileContactSMS, UserProfileContactEmail, UserProfileContactURL, UserProfileContactOther:
		return true
	default:
		return false
	}
}

// UserProfileContact describes one opt-in profile-visible contact method.
//
// URI is the actionable value for clients, such as tel:+15035551212,
// sms:+15035551212, mailto:driver@example.com, or an https URL.
type UserProfileContact struct {
	Kind        UserProfileContactKind `json:"kind"`
	Label       string                 `json:"label,omitempty"`
	DisplayText string                 `json:"display_text,omitempty"`
	URI         string                 `json:"uri"`
	Verified    bool                   `json:"verified"`
}

// Validate reports whether user contains the required identity and profile
// fields.
func (user User) Validate() error {
	if !user.ID.Valid() {
		return errors.New("user id must be a valid UUID")
	}
	if user.DeletionAfterInactivitySeconds != nil && *user.DeletionAfterInactivitySeconds <= 0 {
		return errors.New("deletion_after_inactivity_seconds must be positive")
	}
	if err := user.Profile.Validate(); err != nil {
		return fmt.Errorf("profile: %w", err)
	}
	for i, app := range user.ClientApps {
		if err := app.Validate(); err != nil {
			return fmt.Errorf("client_apps[%d]: %w", i, err)
		}
		if app.UserID != user.ID {
			return fmt.Errorf("client_apps[%d]: user_id does not match user", i)
		}
	}
	return nil
}

// Validate reports whether profile contains a display name and valid optional
// public presentation fields.
func (profile UserProfile) Validate() error {
	if profile.DisplayName == "" {
		return errors.New("display_name must be set")
	}
	if profile.AvatarURL != "" && !validAbsoluteURL(profile.AvatarURL) {
		return errors.New("avatar_url must be an absolute URL")
	}
	if profile.BannerURL != "" && !validAbsoluteURL(profile.BannerURL) {
		return errors.New("banner_url must be an absolute URL")
	}
	if profile.AccentColor != "" && !validHexColor(profile.AccentColor) {
		return errors.New("accent_color must be #RRGGBB")
	}
	for i, link := range profile.Links {
		if err := link.Validate(); err != nil {
			return fmt.Errorf("links[%d]: %w", i, err)
		}
	}
	for i, contact := range profile.Contacts {
		if err := contact.Validate(); err != nil {
			return fmt.Errorf("contacts[%d]: %w", i, err)
		}
	}
	return nil
}

// Validate reports whether link contains an absolute URL.
func (link UserProfileLink) Validate() error {
	if link.URL == "" {
		return errors.New("url must be set")
	}
	if !validAbsoluteURL(link.URL) {
		return errors.New("url must be an absolute URL")
	}
	return nil
}

// Validate reports whether contact contains a known kind and actionable URI.
func (contact UserProfileContact) Validate() error {
	if !contact.Kind.Valid() {
		return errors.New("kind must be a known OpenCaravan value")
	}
	u, err := url.Parse(contact.URI)
	if err != nil || u.Scheme == "" {
		return errors.New("uri must be an absolute URI")
	}
	switch contact.Kind {
	case UserProfileContactPhone:
		if u.Scheme != "tel" {
			return errors.New("phone contact uri must use tel scheme")
		}
	case UserProfileContactSMS:
		if u.Scheme != "sms" {
			return errors.New("sms contact uri must use sms scheme")
		}
	case UserProfileContactEmail:
		if u.Scheme != "mailto" {
			return errors.New("email contact uri must use mailto scheme")
		}
	case UserProfileContactURL:
		if u.Scheme != "http" && u.Scheme != "https" {
			return errors.New("url contact uri must use http or https scheme")
		}
	}
	return nil
}

func validAbsoluteURL(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func validHexColor(value string) bool {
	if len(value) != 7 || value[0] != '#' {
		return false
	}
	for _, r := range value[1:] {
		if !isHex(r) {
			return false
		}
	}
	return true
}
