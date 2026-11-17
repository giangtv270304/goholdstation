package goholdstation

import (
	"bytes"
	"encoding/json"
	"errors"
)

// Response is the result of every successful call.
// Message is the envelope "message" ("success"); it is empty when the API answers without the {message, data} envelope.
type Response[T any] struct {
	Message string `json:"message"`
	Data    T      `json:"data"`
}

// BusinessResult is the business outcome some operations carry next to their payload.
// A false Success is returned as an *APIError, except by GetOrderCustomerInfo.
type BusinessResult struct {
	Success bool `json:"success"`
	// Code is 0 on success. The documented values depend on the operation.
	Code  int    `json:"code"`
	Error string `json:"error"`
}

var errEmptyResponse = errors.New("response body is empty or null")

func splitEnvelope(raw []byte, dataKeyInPayload bool) (string, []byte, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return "", nil, errEmptyResponse
	}

	var top map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &top); err != nil {
		return "", trimmed, nil
	}

	data, wrapped := top["data"]
	if !wrapped {
		return "", trimmed, nil
	}

	if dataKeyInPayload && !wrapsPayloadWithDataKey(top, data) {
		return "", trimmed, nil
	}

	var message string
	_ = json.Unmarshal(top["message"], &message)

	return message, data, nil
}

func wrapsPayloadWithDataKey(top map[string]json.RawMessage, data json.RawMessage) bool {
	for _, key := range []string{"error", "error_code"} {
		if _, ok := top[key]; ok {
			return false
		}
	}

	var inner map[string]json.RawMessage
	if err := json.Unmarshal(data, &inner); err != nil {
		return false
	}

	for _, key := range []string{"data", "error", "error_code", "message"} {
		if _, ok := inner[key]; ok {
			return true
		}
	}

	return false
}

func businessFailure(statusCode int, raw, payload []byte) *APIError {
	var result struct {
		Success *bool           `json:"success"`
		Code    json.RawMessage `json:"code"`
		Error   json.RawMessage `json:"error"`
	}

	if err := json.Unmarshal(payload, &result); err != nil {
		return nil
	}

	var message string
	_ = json.Unmarshal(result.Error, &message)

	code := rawToString(result.Code)

	failed := result.Success != nil && !*result.Success
	if result.Success == nil && message != "" && code != "" && code != "0" {
		failed = true
	}

	if !failed {
		return nil
	}

	if message == "" {
		message = "the API reported success=false"
	}

	return &APIError{StatusCode: statusCode, Code: code, Message: message, Detail: "", Details: nil, Body: raw}
}

func rawToString(raw json.RawMessage) string {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return ""
	}

	if trimmed[0] == '"' {
		var decoded string
		if err := json.Unmarshal(trimmed, &decoded); err == nil {
			return decoded
		}
	}

	return string(trimmed)
}

func firstRaw(candidates ...json.RawMessage) json.RawMessage {
	for _, candidate := range candidates {
		if len(candidate) > 0 && !bytes.Equal(candidate, []byte("null")) {
			return candidate
		}
	}

	return nil
}

func isJSONArray(raw []byte) bool {
	trimmed := bytes.TrimSpace(raw)

	return len(trimmed) > 0 && trimmed[0] == '['
}
