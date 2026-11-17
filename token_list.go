package goholdstation

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
)

// TokenRate is the VND rate of a token, present only when the list was asked WithRate.
type TokenRate struct {
	Buy  Decimal `json:"buy"`
	Sell Decimal `json:"sell"`
}

// Token is a crypto token supported by Holdstation Pay.
type Token struct {
	// ID is a UUID.
	ID           string `json:"id"`
	ChainID      int64  `json:"chain_id"`
	Address      string `json:"address"`
	Name         string `json:"name"`
	Symbol       string `json:"symbol"`
	NameOverride string `json:"name_override"`
	Decimals     int32  `json:"decimals"`
	// Price and PriceChange24h are decimal strings.
	Price              string     `json:"price"`
	PriceChange24h     string     `json:"price_change_24h"`
	Image              string     `json:"image"`
	Active             bool       `json:"active"`
	WithdrawalContract string     `json:"withdrawal_contract"`
	Rate               *TokenRate `json:"rate"`
	CreatedAt          *Timestamp `json:"created_at"`
	UpdatedAt          *Timestamp `json:"updated_at"`
}

// ListTokensRequest: zero values are left out and the API applies its defaults.
type ListTokensRequest struct {
	ChainID int64
	// Active filters on the active flag; nil sends no filter.
	Active   *bool
	Page     int
	PageSize int
	Symbols  []string
	NoLimit  bool
	// WithRate fills Token.Rate.
	WithRate bool
}

// ListTokensResponse: the paging fields are zero when the API answers with a plain array.
type ListTokensResponse struct {
	Tokens     []Token `json:"tokens"`
	Page       int     `json:"page"`
	PageSize   int     `json:"page_size"`
	TotalCount int     `json:"total_count"`
}

// UnmarshalJSON accepts the documented object and the documented plain array of tokens.
func (r *ListTokensResponse) UnmarshalJSON(data []byte) error {
	if isJSONArray(data) {
		return json.Unmarshal(data, &r.Tokens)
	}

	type plain ListTokensResponse

	return json.Unmarshal(data, (*plain)(r))
}

// ListTokens returns one page of tokens. Pass nil for the first page with defaults.
func (c *Client) ListTokens(ctx context.Context, req *ListTokensRequest) (*Response[ListTokensResponse], error) {
	query := url.Values{}

	if req != nil {
		setPositive(query, "chain_id", req.ChainID)

		if req.Active != nil {
			query.Set("active", strconv.FormatBool(*req.Active))
		}

		setPositive(query, "page", int64(req.Page))
		setPositive(query, "page_size", int64(req.PageSize))
		setStringList(query, "symbols", req.Symbols)
		setTrue(query, "no_limit", req.NoLimit)
		setTrue(query, "with_rate", req.WithRate)
	}

	return call[ListTokensResponse](ctx, c, request{
		Method: http.MethodGet,
		Path:   "/api/v1/tokens",
		Query:  query,
		Body:   nil,
	})
}
