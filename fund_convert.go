package goholdstation

import (
	"context"
	"net/http"
)

type ConvertFundRequest struct {
	// FromToken and ToToken are "USDT", "USDC" or "USD1".
	FromToken string `json:"from_token"`
	ToToken   string `json:"to_token"`
	// Amount is a decimal string; empty converts the whole balance.
	Amount         string `json:"amount,omitempty"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

// ConvertFundResponse: Rate, ConvertedAmount, ReceivedAmount and Fee are decimal strings. Rate is always "1".
type ConvertFundResponse struct {
	BusinessResult

	Rate            string   `json:"rate"`
	ConvertedAmount string   `json:"converted_amount"`
	ReceivedAmount  string   `json:"received_amount"`
	Fee             string   `json:"fee"`
	FromFund        *Fund    `json:"from_fund"`
	ToFund          *Fund    `json:"to_fund"`
	FromLog         *FundLog `json:"from_log"`
	ToLog           *FundLog `json:"to_log"`
}

// ConvertFund moves a prefunded balance from one stablecoin to another. Signed. It moves money.
// A business failure is an *APIError with status 200.
// Any error other than an *APIError with a 4xx status or a business code means the outcome is unknown:
// check ListFundLogs, or resend with the same IdempotencyKey, never a new one.
func (c *Client) ConvertFund(ctx context.Context, req *ConvertFundRequest) (*Response[ConvertFundResponse], error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	return call[ConvertFundResponse](ctx, c, request{
		Method: http.MethodPost,
		Path:   "/partners/funds/convert",
		Query:  nil,
		Body:   req,
		Signed: true,
	})
}
