package opencaravan_test

import (
	"strings"
	"testing"

	opencaravan "github.com/opencaravan/opencaravan-go"
)

func TestBlobRefValidate(t *testing.T) {
	good := opencaravan.BlobRef{
		Hash:        "sha256:" + strings.Repeat("a", 64),
		Size:        1024,
		ContentType: "image/png",
	}
	if err := good.Validate(); err != nil {
		t.Fatalf("happy path: %v", err)
	}
	cases := map[string]func(*opencaravan.BlobRef){
		"missing hash":                    func(b *opencaravan.BlobRef) { b.Hash = "" },
		"hash missing prefix":             func(b *opencaravan.BlobRef) { b.Hash = strings.Repeat("a", 64) },
		"hash wrong length":               func(b *opencaravan.BlobRef) { b.Hash = "sha256:" + strings.Repeat("a", 32) },
		"hash upper-case hex":             func(b *opencaravan.BlobRef) { b.Hash = "sha256:" + strings.Repeat("A", 64) },
		"hash non-hex":                    func(b *opencaravan.BlobRef) { b.Hash = "sha256:" + strings.Repeat("z", 64) },
		"negative size":                   func(b *opencaravan.BlobRef) { b.Size = -1 },
		"missing content type":            func(b *opencaravan.BlobRef) { b.ContentType = "" },
		"content type with upper case":    func(b *opencaravan.BlobRef) { b.ContentType = "Image/JPEG" },
		"content type with leading space": func(b *opencaravan.BlobRef) { b.ContentType = " image/jpeg" },
		"content type with trailing tab":  func(b *opencaravan.BlobRef) { b.ContentType = "image/jpeg\t" },
	}
	for name, mut := range cases {
		t.Run(name, func(t *testing.T) {
			b := good
			mut(&b)
			if err := b.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestValidateCanonicalHash(t *testing.T) {
	// A zero-size blob is structurally fine (some clients may legitimately
	// store empty notes/files). The hash itself must still be the canonical
	// shape, which is what this helper polices.
	if err := opencaravan.ValidateCanonicalHash("sha256:" + strings.Repeat("0", 64)); err != nil {
		t.Fatalf("all-zeros canonical hash should validate: %v", err)
	}
	bad := []string{
		"",
		"sha256:",
		"sha1:" + strings.Repeat("a", 40),
		"sha256:" + strings.Repeat("a", 63),
		"sha256:" + strings.Repeat("A", 64),
		"SHA256:" + strings.Repeat("a", 64),
		strings.Repeat("a", 64),
		"sha256:" + strings.Repeat("g", 64),
	}
	for _, s := range bad {
		t.Run(s, func(t *testing.T) {
			if err := opencaravan.ValidateCanonicalHash(s); err == nil {
				t.Fatalf("expected %q to be rejected", s)
			}
		})
	}
}
