package goholdstation

import (
	"context"
	"net/http"
)

type CreateDepositOrderRequest struct {
	// Amount is a decimal string in Currency.
	Amount       string `json:"amount"`
	Currency     string `json:"currency"`
	ChainID      int64  `json:"chain_id"`
	TokenAddress string `json:"token_address"`
	// Recipient is the wallet that receives the tokens.
	Recipient string `json:"recipient"`
	// Callback is the HTTPS URL that receives the order status webhook.
	Callback string `json:"callback"`
	// ReferenceNo must be unique per partner.
	ReferenceNo string `json:"reference_no,omitempty"`
	// Buffer overrides the buffer rate as a fraction ("0.005" is 0.5 percent); nil leaves it out.
	Buffer *Decimal `json:"buffer,omitempty"`
	KYCID  string   `json:"kyc_id,omitempty"`
}

type CreateDepositOrderResponse struct {
	Order *Order `json:"order"`
}

// CreateDepositOrder creates an onramp order: the user pays VND and receives tokens.
// Order.PayData carries the bank account and QR code to pay. A refusal is an *APIError with code "400105" or "400110".
// Any error other than an *APIError with a 4xx status or a business code means the outcome is unknown:
// read the order with GetOrderByReferenceNo, or resend with the same ReferenceNo, never a new one.
func (c *Client) CreateDepositOrder(ctx context.Context, req *CreateDepositOrderRequest) (*Response[CreateDepositOrderResponse], error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	return call[CreateDepositOrderResponse](ctx, c, request{
		Method: http.MethodPost,
		Path:   "/partners/orders/deposit",
		Query:  nil,
		Body:   req,
	})
}
