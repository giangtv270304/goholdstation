package goholdstation

import (
	"context"
	"net/http"
)

type ListMIDVAsRequest struct{}

type ListMIDVAsResponse struct {
	VAs        []MIDVA `json:"vas"`
	TotalCount int     `json:"total_count"`
	Page       int     `json:"page"`
	PageSize   int     `json:"page_size"`
}

// ListMIDVAs returns the merchant virtual accounts, newest first. Signed. Pass nil.
func (c *Client) ListMIDVAs(ctx context.Context, _ *ListMIDVAsRequest) (*Response[ListMIDVAsResponse], error) {
	return call[ListMIDVAsResponse](ctx, c, request{
		Method: http.MethodGet,
		Path:   "/partners/mid/va",
		Query:  nil,
		Body:   nil,
		Signed: true,
	})
}
