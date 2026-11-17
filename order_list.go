package goholdstation

import (
	"context"
	"net/http"
	"net/url"
)

// ListOrdersRequest: zero values are left out and the API applies its defaults.
type ListOrdersRequest struct {
	Types            []OrderType
	States           []OrderState
	ProcessingStates []OrderProcessingState
	Page             int
	PageSize         int
	// SortBy is "created_at", "expired_at" or "amount".
	SortBy   string
	SortDesc bool
	// The four time bounds are Unix seconds.
	CreatedAtFrom int64
	CreatedAtTo   int64
	ExpiredAtFrom int64
	ExpiredAtTo   int64
	Recipient     string
	ReferenceNos  []string
}

type ListOrdersResponse struct {
	Orders     []Order `json:"orders"`
	TotalCount int     `json:"total_count"`
}

// ListOrders returns one page of orders. Pass nil for the first page with defaults.
func (c *Client) ListOrders(ctx context.Context, req *ListOrdersRequest) (*Response[ListOrdersResponse], error) {
	query := url.Values{}

	if req != nil {
		setIntList(query, "type", req.Types)
		setIntList(query, "state", req.States)
		setIntList(query, "processing_state", req.ProcessingStates)
		setPositive(query, "page", int64(req.Page))
		setPositive(query, "page_size", int64(req.PageSize))
		setNonEmpty(query, "sort_by", req.SortBy)
		setTrue(query, "sort_desc", req.SortDesc)
		setPositive(query, "created_at_from", req.CreatedAtFrom)
		setPositive(query, "created_at_to", req.CreatedAtTo)
		setPositive(query, "expired_at_from", req.ExpiredAtFrom)
		setPositive(query, "expired_at_to", req.ExpiredAtTo)
		setNonEmpty(query, "recipient", req.Recipient)
		setStringList(query, "reference_nos", req.ReferenceNos)
	}

	return call[ListOrdersResponse](ctx, c, request{
		Method: http.MethodGet,
		Path:   "/partners/orders",
		Query:  query,
		Body:   nil,
	})
}
