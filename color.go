package opencaravan

import (
	"errors"
	"fmt"
	"strings"
)

// HexColor is an opaque sRGB color encoded as a canonical #rrggbb string.
//
// Accent colors intentionally do not support alpha or alternate color spaces.
// Future protocol fields can add those semantics without changing the meaning
// of this type.
type HexColor string

// ParseHexColor parses and canonicalizes a #RRGGBB color string.
func ParseHexColor(s string) (HexColor, error) {
	color := HexColor(strings.ToLower(s))
	if !color.Valid() {
		return "", fmt.Errorf("invalid hex color %q", s)
	}
	return color, nil
}

// Valid reports whether color is a #RRGGBB color string.
func (color HexColor) Valid() bool {
	value := string(color)
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

// String returns the raw color string.
func (color HexColor) String() string {
	return string(color)
}

// MarshalText returns the canonical lowercase #rrggbb form of color.
func (color HexColor) MarshalText() ([]byte, error) {
	parsed, err := ParseHexColor(string(color))
	if err != nil {
		return nil, err
	}
	return []byte(parsed), nil
}

// UnmarshalText parses a canonical hex color into color.
func (color *HexColor) UnmarshalText(text []byte) error {
	if color == nil {
		return errors.New("unmarshal HexColor into nil pointer")
	}

	parsed, err := ParseHexColor(string(text))
	if err != nil {
		return err
	}
	*color = parsed
	return nil
}
