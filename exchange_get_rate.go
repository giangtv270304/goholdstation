package goholdstation

import (
	"context"
	"net/http"
	"net/url"
)

// ExchangeRate is the USD/VND rate. Buy and Sell are decimal strings ("25850").
type ExchangeRate struct {
	Buy  string `json:"buy"`
	Sell string `json:"sell"`
}

type GetExchangeRateRequest struct {
	// Original asks for the rate before fees are applied.
	Original bool
}

type GetExchangeRateResponse struct {
	Rate *ExchangeRate `json:"rate"`
}

// GetExchangeRate returns the current USD/VND rate. Pass nil for the rate with fees.
func (c *Client) GetExchangeRate(ctx context.Context, req *GetExchangeRateRequest) (*Response[GetExchangeRateResponse], error) {
	query := url.Values{}

	if req != nil {
		setTrue(query, "original", req.Original)
	}

	return call[GetExchangeRateResponse](ctx, c, request{
		Method: http.MethodGet,
		Path:   "/api/v1/exchange/rate",
		Query:  query,
		Body:   nil,
	})
}
