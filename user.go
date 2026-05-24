package opencaravan

import (
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"strings"
)

// User describes a server-scoped OpenCaravan identity.
//
// Servers assign user IDs per connection or registration. A person may have
// different user IDs and different profile details on different servers.
type User struct {
	ID      UUID        `json:"id"`
	Profile UserProfile `json:"profile"`
	// DeletionAfterInactivityDays is an optional duration after which the
	// server may delete this user record if no activity resets the timer.
	DeletionAfterInactivityDays *int64      `json:"deletion_after_inactivity_days,omitempty"`
	ClientApps                  []ClientApp `json:"client_apps,omitempty"`
}

// UserProfile describes client-supplied profile information a server may
// republish to authorized journey participants.
//
// Contacts are explicitly profile-visible. Private account recovery fields,
// billing data, and server-only authentication metadata do not belong here.
// Clients may mirror one profile across servers or tailor profile details for
// each server registration.
type UserProfile struct {
	DisplayName string `json:"display_name"`
	// AvatarImage is the image clients can use for compact or map
	// representations of this user.
	AvatarImage *ImageResourceRef `json:"avatar_image,omitempty"`
	// BannerImage is an optional wide image clients can use in richer profile
	// views.
	BannerImage *ImageResourceRef    `json:"banner_image,omitempty"`
	Bio         string               `json:"bio,omitempty"`
	AccentColor HexColor             `json:"accent_color,omitempty"`
	Links       []UserProfileLink    `json:"links,omitempty"`
	Contacts    []UserProfileContact `json:"contacts,omitempty"`
}

// UserProfileLink describes one user-supplied public profile link.
type UserProfileLink struct {
	Kind  string `json:"kind,omitempty"`
	Label string `json:"label,omitempty"`
	URL   string `json:"url"`
}

const (
	// UserProfileContactMobileNumber is a mobile telephone number that clients
	// may use for compatible local calling or messaging capabilities.
	UserProfileContactMobileNumber = "mobile_number"
	// UserProfileContactEmailAddress is an email address that clients may use
	// with compatible local messaging capabilities.
	UserProfileContactEmailAddress = "email_address"
)

// UserProfileContact describes one opt-in profile-visible contact identifier.
//
// Contacts are addressable identifiers such as a mobile number or email
// address. Links remain separate because they describe public destinations
// rather than direct contact channels.
type UserProfileContact struct {
	Kind        string `json:"kind"`
	Label       string `json:"label,omitempty"`
	Value       string `json:"value"`
	DisplayText string `json:"display_text,omitempty"`
	Verified    bool   `json:"verified"`
}

// Validate reports whether user contains the required identity and profile
// fields.
func (user User) Validate() error {
	if !user.ID.Valid() {
		return errors.New("user id must be a valid UUID")
	}
	if user.DeletionAfterInactivityDays != nil && *user.DeletionAfterInactivityDays <= 0 {
		return errors.New("deletion_after_inactivity_days must be positive")
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
	if profile.AvatarImage != nil {
		if err := profile.AvatarImage.Validate(); err != nil {
			return fmt.Errorf("avatar_image: %w", err)
		}
	}
	if profile.BannerImage != nil {
		if err := profile.BannerImage.Validate(); err != nil {
			return fmt.Errorf("banner_image: %w", err)
		}
	}
	if profile.AccentColor != "" && !profile.AccentColor.Valid() {
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

// Validate reports whether contact contains a kind and addressable value.
func (contact UserProfileContact) Validate() error {
	if strings.TrimSpace(contact.Kind) == "" {
		return errors.New("kind must be set")
	}
	if strings.TrimSpace(contact.Value) == "" {
		return errors.New("value must be set")
	}

	switch contact.Kind {
	case UserProfileContactMobileNumber:
		if !validMobileNumber(contact.Value) {
			return errors.New("mobile_number value must be an E.164-style number")
		}
	case UserProfileContactEmailAddress:
		address, err := mail.ParseAddress(contact.Value)
		if err != nil || address.Name != "" || address.Address != contact.Value {
			return errors.New("email_address value must be an email address")
		}
	}
	return nil
}

func validAbsoluteURL(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func validMobileNumber(value string) bool {
	if len(value) < 3 || len(value) > 16 || value[0] != '+' {
		return false
	}
	for _, r := range value[1:] {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
