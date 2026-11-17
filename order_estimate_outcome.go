package goholdstation

import (
	"context"
	"net/http"
)

type EstimateOrderOutcomeRequest struct {
	// Amount is a decimal string.
	Amount       string    `json:"amount"`
	OrderType    OrderType `json:"order_type"`
	ChainID      int64     `json:"chain_id"`
	TokenAddress string    `json:"token_address"`
	Reverse      bool      `json:"reverse"`
	// Buffer overrides the buffer rate as a fraction ("0.005" is 0.5 percent); nil leaves it out.
	Buffer *Decimal `json:"buffer,omitempty"`
}

type EstimateOrderOutcomeResponse struct {
	// Outcome is a decimal string: tokens (USD) for onramp, VND for offramp.
	Outcome string      `json:"outcome"`
	Rate    Decimal     `json:"rate"`
	FeeVND  Decimal     `json:"fee_vnd"`
	Fees    *AppliedFee `json:"fees"`
}

// EstimateOrderOutcome quotes what an order of the given amount would produce. It creates nothing.
func (c *Client) EstimateOrderOutcome(ctx context.Context, req *EstimateOrderOutcomeRequest) (*Response[EstimateOrderOutcomeResponse], error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	return call[EstimateOrderOutcomeResponse](ctx, c, request{
		Method: http.MethodPost,
		Path:   "/api/v1/payments/estimate-outcome",
		Query:  nil,
		Body:   req,
	})
}
