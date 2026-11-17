package goholdstation

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// Values of DepositVA.Status.
const (
	DepositVAStatusActive = "ACTIVE"
	DepositVAStatusClosed = "CLOSED"
)

// DepositVABank is the bank that holds a VND deposit account.
type DepositVABank struct {
	ID        int64  `json:"id"`
	BinCode   string `json:"bin_code"`
	ShortName string `json:"short_name"`
	BankName  string `json:"bank_name"`
	Image     string `json:"image"`
}

// DepositVA is a VND bank account that tops up the prefunded VND balance.
type DepositVA struct {
	// ID looks like "dva_01HQ8X2M4K".
	ID            string         `json:"id"`
	AccountNumber string         `json:"account_number"`
	AccountName   string         `json:"account_name"`
	BankID        int64          `json:"bank_id"`
	Bank          *DepositVABank `json:"bank"`
	Status        string         `json:"status"`
	// CreatedAt is an RFC 3339 string.
	CreatedAt string `json:"created_at"`
}

type GetDepositVAByIDRequest struct {
	VAID string
}

type GetDepositVAByIDResponse = DepositVA

// GetDepositVAByID returns one VND deposit account. Signed. An unknown id is a 404 *APIError.
// It fails with ErrMissingField when VAID is empty.
func (c *Client) GetDepositVAByID(ctx context.Context, req *GetDepositVAByIDRequest) (*Response[GetDepositVAByIDResponse], error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	if req.VAID == "" {
		return nil, fmt.Errorf("%w: VAID", ErrMissingField)
	}

	return call[GetDepositVAByIDResponse](ctx, c, request{
		Method: http.MethodGet,
		Path:   "/partners/funds/deposit-vas/" + url.PathEscape(req.VAID),
		Query:  nil,
		Body:   nil,
		Signed: true,
	})
}
