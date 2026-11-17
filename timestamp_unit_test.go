package goholdstation_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/giangtv270304/goholdstation"
)

func TestTimestamp_DecodesBothDocumentedForms(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		raw  string
		want goholdstation.Timestamp
	}{
		"object":            {`{"seconds":1700000000,"nanos":5}`, goholdstation.Timestamp{Seconds: 1700000000, Nanos: 5}},
		"object extra keys": {`{"seconds":1700000000,"nanos":0,"zone":"x"}`, goholdstation.Timestamp{Seconds: 1700000000, Nanos: 0}},
		"rfc 3339":          {`"2026-10-29T03:14:22Z"`, goholdstation.Timestamp{Seconds: 1793243662, Nanos: 0}},
		"rfc 3339 offset":   {`"2026-10-29T10:14:22.5+07:00"`, goholdstation.Timestamp{Seconds: 1793243662, Nanos: 500000000}},
		"empty string":      {`""`, goholdstation.Timestamp{}},
		"null":              {`null`, goholdstation.Timestamp{}},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var decoded goholdstation.Timestamp
			if err := json.Unmarshal([]byte(testCase.raw), &decoded); err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}

			if decoded != testCase.want {
				t.Errorf("decoded = %+v, want %+v", decoded, testCase.want)
			}
		})
	}
}

func TestTimestamp_TimeAndZero(t *testing.T) {
	t.Parallel()

	timestamp := goholdstation.Timestamp{Seconds: 1793243662, Nanos: 0}

	if want := time.Date(2026, 10, 29, 3, 14, 22, 0, time.UTC); !timestamp.Time().Equal(want) {
		t.Errorf("Time() = %v, want %v", timestamp.Time(), want)
	}

	if timestamp.IsZero() || !(goholdstation.Timestamp{}).IsZero() {
		t.Error("IsZero is wrong")
	}
}

func TestTimestamp_RejectsGarbage(t *testing.T) {
	t.Parallel()

	var decoded goholdstation.Timestamp
	for _, raw := range []string{`"yesterday"`, `12`, `[1]`} {
		if err := json.Unmarshal([]byte(raw), &decoded); err == nil {
			t.Errorf("Unmarshal(%s) succeeded, want an error", raw)
		}
	}
}
