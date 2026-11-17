package goholdstation_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestAPIError_NonJSONBodyIsKept(t *testing.T) {
	t.Parallel()

	const html = "<html><body>502 Bad Gateway</body></html>"

	client, _ := newFakeAPI(t, http.StatusBadGateway, html)

	_, err := client.GetExchangeRate(context.Background(), nil)

	apiErr := requireAPIError(t, err, http.StatusBadGateway, "")
	if string(apiErr.Body) != html || apiErr.Message != html {
		t.Errorf("body = %q, message = %q, want the HTML page", apiErr.Body, apiErr.Message)
	}
}

func TestAPIError_FieldsAndMessage(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusUnauthorized, unauthorizedBody)

	_, err := client.GetExchangeRate(context.Background(), nil)

	apiErr := requireAPIError(t, err, http.StatusUnauthorized, "UNAUTHORIZED")
	if apiErr.Message != "Invalid or missing API key" || string(apiErr.Body) != unauthorizedBody {
		t.Errorf("api error = %+v", apiErr)
	}

	want := "goholdstation: api error: status=401 code=UNAUTHORIZED message=Invalid or missing API key"
	if err.Error() != want {
		t.Errorf("Error() = %q, want %q", err.Error(), want)
	}
}

func TestAPIError_CodeFallsBackToErrorAndDetailIsParsed(t *testing.T) {
	t.Parallel()

	const body = `{"error":"VALIDATION_FAILED","message":"One or more fields are invalid","detail":"{\"code\":\"VALIDATION_FAILED\",\"details\":[{\"code\":\"MISSING_FIELD\",\"target\":\"/persons/0/date_of_birth\"},{\"code\":\"FILE_TOO_LARGE\",\"target\":\"/documents/2/files/0\"}]}"}`

	client, _ := newFakeAPI(t, http.StatusUnprocessableEntity, body)

	_, err := client.SubmitKYB(context.Background(), &goholdstation.SubmitKYBRequest{})

	apiErr := requireAPIError(t, err, http.StatusUnprocessableEntity, "VALIDATION_FAILED")
	if apiErr.Message != "One or more fields are invalid" {
		t.Errorf("message = %q", apiErr.Message)
	}

	want := []goholdstation.ErrorDetail{
		{Code: "MISSING_FIELD", Target: "/persons/0/date_of_birth"},
		{Code: "FILE_TOO_LARGE", Target: "/documents/2/files/0"},
	}

	if len(apiErr.Details) != len(want) || apiErr.Details[0] != want[0] || apiErr.Details[1] != want[1] {
		t.Errorf("details = %+v, want %+v", apiErr.Details, want)
	}

	if !strings.Contains(apiErr.Detail, "VALIDATION_FAILED") {
		t.Errorf("detail = %q, want the raw detail string", apiErr.Detail)
	}
}

func TestAPIError_StatusHelpers(t *testing.T) {
	t.Parallel()

	helpers := map[int]func(error) bool{
		http.StatusNotFound:        goholdstation.IsNotFound,
		http.StatusUnauthorized:    goholdstation.IsUnauthorized,
		http.StatusConflict:        goholdstation.IsConflict,
		http.StatusTooManyRequests: goholdstation.IsRateLimited,
	}

	for status, matches := range helpers {
		client, _ := newFakeAPI(t, status, badRequestBody)

		_, err := client.GetExchangeRate(context.Background(), nil)
		wrapped := fmt.Errorf("outer: %w", err)

		for otherStatus, helper := range helpers {
			if got, want := helper(wrapped), otherStatus == status; got != want {
				t.Errorf("status %d: helper for %d = %v, want %v", status, otherStatus, got, want)
			}
		}

		if !matches(wrapped) {
			t.Errorf("status %d: helper does not match through a wrapped error", status)
		}
	}

	if goholdstation.IsNotFound(nil) || goholdstation.IsNotFound(context.Canceled) {
		t.Error("IsNotFound matches a non-API error")
	}
}

func TestAPIError_LongNonJSONBodyIsTruncatedInMessageOnly(t *testing.T) {
	t.Parallel()

	page := "<html>" + strings.Repeat("x", 5000) + "</html>"

	client, _ := newFakeAPI(t, http.StatusBadGateway, page)

	_, err := client.GetExchangeRate(context.Background(), nil)

	apiErr := requireAPIError(t, err, http.StatusBadGateway, "")
	if len(apiErr.Message) != 200 || !strings.HasPrefix(page, apiErr.Message) {
		t.Errorf("message has %d bytes, want the first 200 of the body", len(apiErr.Message))
	}

	if string(apiErr.Body) != page {
		t.Errorf("raw body has %d bytes, want %d", len(apiErr.Body), len(page))
	}

	if len(err.Error()) > 400 {
		t.Errorf("Error() has %d bytes, want a short message", len(err.Error()))
	}
}
