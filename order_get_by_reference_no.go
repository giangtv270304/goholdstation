package goholdstation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type GetOrderByReferenceNoRequest struct {
	ReferenceNo string
}

// GetOrderByReferenceNoResponse: the API documents two forms, "order" and "orders".
// Both fill Order (the first match) and Orders.
type GetOrderByReferenceNoResponse struct {
	Order    *Order         `json:"order"`
	Orders   []Order        `json:"orders"`
	Payments []OrderPayment `json:"payments"`
	// TotalReceived is a decimal string.
	TotalReceived string `json:"total_received"`
	TotalCount    int    `json:"total_count"`
}

// UnmarshalJSON accepts {"order": {...}} and {"orders": [...], "total_count": n}.
func (r *GetOrderByReferenceNoResponse) UnmarshalJSON(data []byte) error {
	type plain GetOrderByReferenceNoResponse

	if err := json.Unmarshal(data, (*plain)(r)); err != nil {
		return err
	}

	if r.Order == nil && len(r.Orders) > 0 {
		first := r.Orders[0]
		r.Order = &first
	}

	if r.Order != nil && len(r.Orders) == 0 {
		r.Orders = []Order{*r.Order}
	}

	return nil
}

// GetOrderByReferenceNo returns the order created with the partner's reference number.
// It fails with ErrMissingField when ReferenceNo is empty.
func (c *Client) GetOrderByReferenceNo(ctx context.Context, req *GetOrderByReferenceNoRequest) (*Response[GetOrderByReferenceNoResponse], error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	if req.ReferenceNo == "" {
		return nil, fmt.Errorf("%w: ReferenceNo", ErrMissingField)
	}

	return call[GetOrderByReferenceNoResponse](ctx, c, request{
		Method: http.MethodGet,
		Path:   "/partners/orders/reference-no/" + url.PathEscape(req.ReferenceNo),
		Query:  nil,
		Body:   nil,
	})
}
