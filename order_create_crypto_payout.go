package goholdstation

import (
	"context"
	"net/http"
)

// CryptoPayoutCodeUnderMaintenance is the business code of CreateCryptoPayoutOrder during a maintenance window.
const CryptoPayoutCodeUnderMaintenance = 12

type CreateCryptoPayoutOrderRequest struct {
	// Amount is a decimal string in Currency ("vnd", the default, or "token").
	Amount   string `json:"amount"`
	Currency string `json:"currency,omitempty"`
	// TokenAddress must be a stablecoin.
	TokenAddress string `json:"token_address"`
	ChainID      int64  `json:"chain_id"`
	// Recipient is the wallet that receives the tokens.
	Recipient   string `json:"recipient"`
	Callback    string `json:"callback,omitempty"`
	ReferenceNo string `json:"reference_no,omitempty"`
	// IdempotencyKey: sending the same key again returns the original order.
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

type CreateCryptoPayoutOrderResponse struct {
	BusinessResult

	Order *Order `json:"order"`
}

// CreateCryptoPayoutOrder pays stablecoins to a wallet from the prefunded VND balance. Signed.
// A business failure is an *APIError with status 200 and the code in Code.
// Any error other than an *APIError with a 4xx status or a business code means the outcome is unknown:
// resend with the same IdempotencyKey (it returns the original order), never a new one.
func (c *Client) CreateCryptoPayoutOrder(ctx context.Context, req *CreateCryptoPayoutOrderRequest) (*Response[CreateCryptoPayoutOrderResponse], error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	return call[CreateCryptoPayoutOrderResponse](ctx, c, request{
		Method: http.MethodPost,
		Path:   "/partners/orders/crypto-payout",
		Query:  nil,
		Body:   req,
		Signed: true,
	})
}
