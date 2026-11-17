package goholdstation_test

import (
	"encoding/json"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestDecimal_RoundTripKeepsEveryDigit(t *testing.T) {
	t.Parallel()

	for _, digits := range []string{
		"0.1",
		"0.005",
		"25850",
		"12345678901234567890.123456789012345678",
		"0.30000000000000004000",
		"-7.50",
		"1e-7",
	} {
		var decoded struct {
			Value goholdstation.Decimal `json:"value"`
		}

		if err := json.Unmarshal([]byte(`{"value":`+digits+`}`), &decoded); err != nil {
			t.Fatalf("%s: Unmarshal: %v", digits, err)
		}

		if decoded.Value.String() != digits {
			t.Errorf("decoded = %q, want %q", decoded.Value, digits)
		}

		encoded, err := json.Marshal(decoded)
		if err != nil {
			t.Fatalf("%s: Marshal: %v", digits, err)
		}

		if want := `{"value":` + digits + `}`; string(encoded) != want {
			t.Errorf("encoded = %s, want %s", encoded, want)
		}
	}
}

func TestDecimal_UnmarshalFromStringAndNull(t *testing.T) {
	t.Parallel()

	var decoded struct {
		FromString goholdstation.Decimal  `json:"from_string"`
		FromNull   goholdstation.Decimal  `json:"from_null"`
		Pointer    *goholdstation.Decimal `json:"pointer"`
	}

	body := `{"from_string":"1000000.000000000000000001","from_null":null,"pointer":null}`
	if err := json.Unmarshal([]byte(body), &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.FromString != "1000000.000000000000000001" || decoded.FromNull != "" || decoded.Pointer != nil {
		t.Errorf("decoded = %+v", decoded)
	}

	encoded, err := json.Marshal(decoded)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	if want := `{"from_string":1000000.000000000000000001,"from_null":null,"pointer":null}`; string(encoded) != want {
		t.Errorf("encoded = %s, want %s", encoded, want)
	}
}

func TestDecimal_RejectsWhatIsNotANumber(t *testing.T) {
	t.Parallel()

	for _, value := range []goholdstation.Decimal{"abc", "1,000", "1.2.3", `1,"x":2`, " 1", "true", "{}"} {
		if _, err := json.Marshal(value); err == nil {
			t.Errorf("Marshal(%q) succeeded, want an error", value)
		}
	}

	var decoded goholdstation.Decimal
	for _, raw := range []string{`true`, `{}`, `[1]`} {
		if err := json.Unmarshal([]byte(raw), &decoded); err == nil {
			t.Errorf("Unmarshal(%s) succeeded, want an error", raw)
		}
	}
}
