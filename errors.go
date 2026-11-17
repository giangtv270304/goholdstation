package goholdstation

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// APIError is returned for every response with status 400 or above, and for an HTTP 2xx
// response that reports a business failure (success=false). Read it with errors.As.
// The API documents no request id.
type APIError struct {
	// StatusCode is the HTTP status; 200 for a business failure.
	StatusCode int
	// Code is "error_code" (or "error" when the API sends no error_code), or the numeric business code as a string.
	Code string
	// Message is the API message. For a body that is not JSON it is the first 200 bytes of the body.
	Message string
	// Detail is the raw "detail" string; on 422 it is a JSON document, parsed into Details.
	Detail  string
	Details []ErrorDetail
	// Body is the raw response body, kept even when it is not JSON.
	Body []byte
}

// ErrorDetail is one validation failure. Target is a JSON pointer into the request body, empty for the whole body.
type ErrorDetail struct {
	Code   string `json:"code"`
	Target string `json:"target"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("goholdstation: api error: status=%d code=%s message=%s", e.StatusCode, e.Code, e.Message)
}

const maxRawMessageBytes = 200

type errorBody struct {
	Error     string `json:"error"`
	ErrorCode string `json:"error_code"`
	Message   string `json:"message"`
	Detail    string `json:"detail"`
}

func newAPIError(statusCode int, raw []byte) *APIError {
	apiErr := &APIError{StatusCode: statusCode, Code: "", Message: string(raw), Detail: "", Details: nil, Body: raw}
	if len(raw) > maxRawMessageBytes {
		apiErr.Message = string(raw[:maxRawMessageBytes])
	}

	var parsed errorBody
	if err := json.Unmarshal(raw, &parsed); err != nil || (parsed.Error == "" && parsed.ErrorCode == "" && parsed.Message == "") {
		return apiErr
	}

	apiErr.Code = parsed.ErrorCode
	if apiErr.Code == "" {
		apiErr.Code = parsed.Error
	}

	apiErr.Message = parsed.Message
	if apiErr.Message == "" {
		apiErr.Message = parsed.Error
	}

	apiErr.Detail = parsed.Detail

	var detail struct {
		Details []ErrorDetail `json:"details"`
	}

	if err := json.Unmarshal([]byte(parsed.Detail), &detail); err == nil {
		apiErr.Details = detail.Details
	}

	return apiErr
}

func hasStatus(err error, statusCode int) bool {
	var apiErr *APIError

	return errors.As(err, &apiErr) && apiErr.StatusCode == statusCode
}

// IsNotFound reports whether err is an *APIError with status 404.
func IsNotFound(err error) bool { return hasStatus(err, http.StatusNotFound) }

// IsUnauthorized reports whether err is an *APIError with status 401 (API key, signature or timestamp).
func IsUnauthorized(err error) bool { return hasStatus(err, http.StatusUnauthorized) }

// IsConflict reports whether err is an *APIError with status 409.
func IsConflict(err error) bool { return hasStatus(err, http.StatusConflict) }

// IsRateLimited reports whether err is an *APIError with status 429.
func IsRateLimited(err error) bool { return hasStatus(err, http.StatusTooManyRequests) }
