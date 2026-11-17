package goholdstation

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type CloseOrderRequest struct {
	OrderID string
}

type CloseOrderResponse struct {
	BusinessResult
}

// CloseOrder closes an order that is Processing with a pending processing state.
// A refusal (success=false) is an *APIError with status 200. It fails with ErrMissingField when OrderID is empty.
func (c *Client) CloseOrder(ctx context.Context, req *CloseOrderRequest) (*Response[CloseOrderResponse], error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	if req.OrderID == "" {
		return nil, fmt.Errorf("%w: OrderID", ErrMissingField)
	}

	return call[CloseOrderResponse](ctx, c, request{
		Method: http.MethodPost,
		Path:   "/partners/orders/" + url.PathEscape(req.OrderID) + "/close",
		Query:  nil,
		Body:   nil,
	})
}
