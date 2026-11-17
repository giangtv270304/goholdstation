package goholdstation

import (
	"context"
	"net/http"
	"net/url"
)

// Bank is a Vietnamese bank supported for payouts and deposits.
type Bank struct {
	ID        int64  `json:"id"`
	BankName  string `json:"bank_name"`
	ShortName string `json:"short_name"`
	// BinCode is the bank identification number used by VerifyBankAccount.
	BinCode   string     `json:"bin_code"`
	Image     string     `json:"image"`
	CreatedAt *Timestamp `json:"created_at"`
	UpdatedAt *Timestamp `json:"updated_at"`
}

// ListBanksRequest: zero values are left out and the API applies its defaults.
type ListBanksRequest struct {
	Page int
	// Size is the page size (the API names this parameter "size").
	Size int
	IDs  []int64
}

type ListBanksResponse struct {
	Banks      []Bank `json:"banks"`
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
	TotalCount int    `json:"total_count"`
}

// ListBanks returns one page of banks. Pass nil for the first page with defaults.
func (c *Client) ListBanks(ctx context.Context, req *ListBanksRequest) (*Response[ListBanksResponse], error) {
	query := url.Values{}

	if req != nil {
		setPositive(query, "page", int64(req.Page))
		setPositive(query, "size", int64(req.Size))
		setIntList(query, "ids", req.IDs)
	}

	return call[ListBanksResponse](ctx, c, request{
		Method: http.MethodGet,
		Path:   "/api/v1/banks",
		Query:  query,
		Body:   nil,
	})
}
