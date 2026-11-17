package goholdstation

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// Fund types of Fund.FundType.
const (
	FundTypeCrypto = 1
	FundTypeFiat   = 2
)

// Fund is the prefunded balance of one token. Balance is a decimal string: never parse it into a float.
type Fund struct {
	ID        int32  `json:"id"`
	PartnerID string `json:"partner_id"`
	// Token is "USDT", "USDC", "USD1" or "VND".
	Token     string     `json:"token"`
	Balance   string     `json:"balance"`
	FundType  int32      `json:"fund_type"`
	UpdatedAt *Timestamp `json:"updated_at"`
}

type GetFundBalanceRequest struct {
	// Token is required, for example "USDT".
	Token string
}

type GetFundBalanceResponse struct {
	BusinessResult

	Fund *Fund `json:"fund"`
}

// GetFundBalance returns the prefunded balance of one token. Signed.
// It fails with ErrMissingField when Token is empty.
func (c *Client) GetFundBalance(ctx context.Context, req *GetFundBalanceRequest) (*Response[GetFundBalanceResponse], error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	if req.Token == "" {
		return nil, fmt.Errorf("%w: Token", ErrMissingField)
	}

	query := url.Values{}
	query.Set("token", req.Token)

	return call[GetFundBalanceResponse](ctx, c, request{
		Method: http.MethodGet,
		Path:   "/partners/funds/balance",
		Query:  query,
		Body:   nil,
		Signed: true,
	})
}
