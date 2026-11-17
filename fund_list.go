package goholdstation

import (
	"context"
	"net/http"
)

type ListFundsRequest struct{}

type ListFundsResponse struct {
	BusinessResult

	Funds []Fund `json:"funds"`
}

// ListFunds returns every prefunded balance of the partner. Signed. Pass nil.
func (c *Client) ListFunds(ctx context.Context, _ *ListFundsRequest) (*Response[ListFundsResponse], error) {
	return call[ListFundsResponse](ctx, c, request{
		Method: http.MethodGet,
		Path:   "/partners/funds",
		Query:  nil,
		Body:   nil,
		Signed: true,
	})
}
