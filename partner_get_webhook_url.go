package goholdstation

import (
	"context"
	"net/http"
)

type GetWebhookURLRequest struct{}

type GetWebhookURLResponse struct {
	// WebhookURL is empty when no URL is set.
	WebhookURL string `json:"webhook_url"`
}

// GetWebhookURL returns the URL that receives the fund balance webhook. Signed. Pass nil.
func (c *Client) GetWebhookURL(ctx context.Context, _ *GetWebhookURLRequest) (*Response[GetWebhookURLResponse], error) {
	return call[GetWebhookURLResponse](ctx, c, request{
		Method: http.MethodGet,
		Path:   "/partners/webhook-url",
		Query:  nil,
		Body:   nil,
		Signed: true,
	})
}
