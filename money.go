package goholdstation

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Decimal is a money, rate or fee value that the API sends as a JSON number.
// It keeps the digits as text ("0.005", "25850"): it never goes through float64.
// It marshals as a raw JSON number (null when empty) and unmarshals from a JSON number or string.
type Decimal string

// String returns the decimal text.
func (d Decimal) String() string { return string(d) }

// MarshalJSON writes the digits as a JSON number. It fails when the value is not a valid JSON number.
func (d Decimal) MarshalJSON() ([]byte, error) {
	if d == "" {
		return []byte("null"), nil
	}

	if !isJSONNumber([]byte(d)) {
		return nil, fmt.Errorf("goholdstation: Decimal %q is not a valid JSON number", string(d))
	}

	return []byte(d), nil
}

// UnmarshalJSON keeps the digits of a JSON number, or the content of a JSON string, unchanged.
func (d *Decimal) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)

	switch {
	case bytes.Equal(trimmed, []byte("null")):
		*d = ""
	case len(trimmed) > 0 && trimmed[0] == '"':
		var text string
		if err := json.Unmarshal(trimmed, &text); err != nil {
			return fmt.Errorf("goholdstation: decode Decimal: %w", err)
		}

		*d = Decimal(text)
	case isJSONNumber(trimmed):
		*d = Decimal(trimmed)
	default:
		return fmt.Errorf("goholdstation: decode Decimal: %s is not a number or a string", trimmed)
	}

	return nil
}

func isJSONNumber(text []byte) bool {
	if len(text) == 0 || (text[0] != '-' && (text[0] < '0' || text[0] > '9')) {
		return false
	}

	return json.Valid(text)
}
