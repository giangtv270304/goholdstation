package goholdstation

import (
	"context"
	"net/http"
)

type SetWebhookURLRequest struct {
	// WebhookURL must be a public HTTPS URL; an empty string clears it.
	WebhookURL string `json:"webhook_url"`
}

type SetWebhookURLResponse struct {
	WebhookURL string `json:"webhook_url"`
}

// SetWebhookURL sets or clears the URL that receives the fund balance webhook. Signed.
// It does not change the per-order callback. A rejected URL is an *APIError with code "1".
func (c *Client) SetWebhookURL(ctx context.Context, req *SetWebhookURLRequest) (*Response[SetWebhookURLResponse], error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	return call[SetWebhookURLResponse](ctx, c, request{
		Method: http.MethodPost,
		Path:   "/partners/webhook-url",
		Query:  nil,
		Body:   req,
		Signed: true,
	})
}
