package opencaravan

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
)

// CanonicalJSON returns the deterministic JSON encoding of v: object keys
// sorted lexicographically at every level, no insignificant whitespace,
// RFC3339Nano timestamps for time.Time values, fields tagged with
// `,omitempty` omitted when at their zero value.
//
// Different conformant OpenCaravan implementations produce byte-identical
// CanonicalJSON output for the same input value. It is the input bytes
// over which OpenCaravan protocol object signatures are computed; a
// verifier reproduces these bytes and checks the signature against the
// issuing key.
//
// Encoding scheme (see docs/vehicles.md for the worked example): the value
// is first marshaled with encoding/json, parsed back into a generic
// any tree, then re-encoded with sorted keys and no whitespace.
// Numbers round-trip through float64; OpenCaravan integer fields stay
// within the float64 mantissa range by design so no precision is lost.
func CanonicalJSON(v any) ([]byte, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("canonical: marshal: %w", err)
	}
	var generic any
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&generic); err != nil {
		return nil, fmt.Errorf("canonical: round-trip decode: %w", err)
	}
	var buf bytes.Buffer
	if err := encodeCanonical(&buf, generic); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func encodeCanonical(buf *bytes.Buffer, v any) error {
	switch x := v.(type) {
	case nil:
		buf.WriteString("null")
	case bool:
		if x {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
	case string:
		b, err := json.Marshal(x)
		if err != nil {
			return fmt.Errorf("canonical: string: %w", err)
		}
		buf.Write(b)
	case json.Number:
		buf.WriteString(string(x))
	case float64:
		b, err := json.Marshal(x)
		if err != nil {
			return fmt.Errorf("canonical: number: %w", err)
		}
		buf.Write(b)
	case []any:
		buf.WriteByte('[')
		for i, elem := range x {
			if i > 0 {
				buf.WriteByte(',')
			}
			if err := encodeCanonical(buf, elem); err != nil {
				return err
			}
		}
		buf.WriteByte(']')
	case map[string]any:
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		buf.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				buf.WriteByte(',')
			}
			kb, err := json.Marshal(k)
			if err != nil {
				return fmt.Errorf("canonical: key %q: %w", k, err)
			}
			buf.Write(kb)
			buf.WriteByte(':')
			if err := encodeCanonical(buf, x[k]); err != nil {
				return err
			}
		}
		buf.WriteByte('}')
	default:
		return fmt.Errorf("canonical: unsupported type %T", v)
	}
	return nil
}
