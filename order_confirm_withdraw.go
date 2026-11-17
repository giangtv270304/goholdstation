package goholdstation

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type ConfirmWithdrawOrderRequest struct {
	OrderID string
}

type ConfirmWithdrawOrderResponse struct {
	BusinessResult

	Order *Order `json:"order"`
}

// ConfirmWithdrawOrder is step 2 of a prefund withdrawal: it debits the fund and queues the VND payout.
// Signed. It has no idempotency key: after a network error, read the order before calling it again.
// A business failure is an *APIError with status 200. It fails with ErrMissingField when OrderID is empty.
// Any error other than an *APIError with a 4xx status or a business code means the outcome is unknown:
// read the order with GetOrderByID before deciding; do not confirm again blindly.
func (c *Client) ConfirmWithdrawOrder(ctx context.Context, req *ConfirmWithdrawOrderRequest) (*Response[ConfirmWithdrawOrderResponse], error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	if req.OrderID == "" {
		return nil, fmt.Errorf("%w: OrderID", ErrMissingField)
	}

	return call[ConfirmWithdrawOrderResponse](ctx, c, request{
		Method: http.MethodPost,
		Path:   "/partners/orders/" + url.PathEscape(req.OrderID) + "/confirm-withdraw",
		Query:  nil,
		Body:   nil,
		Signed: true,
	})
}
