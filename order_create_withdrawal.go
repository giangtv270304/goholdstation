package goholdstation

import (
	"context"
	"net/http"
)

type CreateWithdrawalOrderRequest struct {
	// Amount is a decimal string in Currency.
	Amount       string       `json:"amount"`
	Currency     string       `json:"currency"`
	ChainID      int64        `json:"chain_id"`
	TokenAddress string       `json:"token_address"`
	PaymentInfo  *PaymentInfo `json:"payment_info"`
	// Callback is the HTTPS URL that receives the order status webhook.
	Callback    string `json:"callback"`
	ReferenceNo string `json:"reference_no,omitempty"`
	// Buffer overrides the buffer rate as a fraction ("0.005" is 0.5 percent); nil leaves it out.
	Buffer *Decimal `json:"buffer,omitempty"`
	KYCID  string   `json:"kyc_id,omitempty"`
	// RefundAddress receives the crypto back if the bank payout is rejected. Not for native tokens.
	RefundAddress string `json:"refund_address,omitempty"`
}

type CreateWithdrawalOrderResponse struct {
	Order *Order `json:"order"`
}

// CreateWithdrawalOrder creates an offramp order: the user sends crypto to Order.Recipient and receives VND.
// Any error other than an *APIError with a 4xx status or a business code means the outcome is unknown:
// read the order with GetOrderByReferenceNo, or resend with the same ReferenceNo, never a new one.
func (c *Client) CreateWithdrawalOrder(ctx context.Context, req *CreateWithdrawalOrderRequest) (*Response[CreateWithdrawalOrderResponse], error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	return call[CreateWithdrawalOrderResponse](ctx, c, request{
		Method: http.MethodPost,
		Path:   "/partners/orders/withdrawal",
		Query:  nil,
		Body:   req,
	})
}
