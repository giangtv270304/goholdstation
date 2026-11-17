package goholdstation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

// Timestamp is a point in time. The API sends it as {"seconds", "nanos"} or as an RFC 3339 string,
// depending on the endpoint; both decode into this type.
type Timestamp struct {
	// Seconds since the Unix epoch.
	Seconds int64 `json:"seconds"`
	Nanos   int32 `json:"nanos"`
}

// Time returns the timestamp in UTC.
func (t Timestamp) Time() time.Time {
	return time.Unix(t.Seconds, int64(t.Nanos)).UTC()
}

// IsZero reports whether the timestamp was absent or empty.
func (t Timestamp) IsZero() bool {
	return t.Seconds == 0 && t.Nanos == 0
}

// UnmarshalJSON accepts {"seconds", "nanos"}, an RFC 3339 string, an empty string or null.
func (t *Timestamp) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)

	if bytes.Equal(trimmed, []byte("null")) {
		*t = Timestamp{Seconds: 0, Nanos: 0}

		return nil
	}

	if len(trimmed) > 0 && trimmed[0] == '"' {
		var text string
		if err := json.Unmarshal(trimmed, &text); err != nil {
			return fmt.Errorf("goholdstation: decode Timestamp: %w", err)
		}

		if text == "" {
			*t = Timestamp{Seconds: 0, Nanos: 0}

			return nil
		}

		parsed, err := time.Parse(time.RFC3339Nano, text)
		if err != nil {
			return fmt.Errorf("goholdstation: decode Timestamp: %w", err)
		}

		*t = Timestamp{Seconds: parsed.Unix(), Nanos: int32(parsed.Nanosecond())}

		return nil
	}

	type plain Timestamp

	var decoded plain
	if err := json.Unmarshal(trimmed, &decoded); err != nil {
		return fmt.Errorf("goholdstation: decode Timestamp: %w", err)
	}

	*t = Timestamp(decoded)

	return nil
}
