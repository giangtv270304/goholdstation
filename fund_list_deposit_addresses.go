package goholdstation

import (
	"context"
	"net/http"
	"net/url"
)

// FundDepositAddress is the address to send crypto to in order to top up the prefunded balance.
type FundDepositAddress struct {
	PartnerID string `json:"partner_id"`
	ChainID   int64  `json:"chain_id"`
	Address   string `json:"address"`
}

type ListFundDepositAddressesRequest struct {
	// ChainID filters on one chain; zero returns every chain.
	ChainID int64
}

type ListFundDepositAddressesResponse struct {
	Success          bool                 `json:"success"`
	DepositAddresses []FundDepositAddress `json:"deposit_addresses"`
	Page             int                  `json:"page"`
	PageSize         int                  `json:"page_size"`
	TotalCount       int                  `json:"total_count"`
}

// ListFundDepositAddresses returns the crypto deposit addresses of the partner. Signed. Pass nil for every chain.
func (c *Client) ListFundDepositAddresses(ctx context.Context, req *ListFundDepositAddressesRequest) (*Response[ListFundDepositAddressesResponse], error) {
	query := url.Values{}

	if req != nil {
		setPositive(query, "chain_id", req.ChainID)
	}

	return call[ListFundDepositAddressesResponse](ctx, c, request{
		Method: http.MethodGet,
		Path:   "/partners/funds/deposit-addresses",
		Query:  query,
		Body:   nil,
		Signed: true,
	})
}
