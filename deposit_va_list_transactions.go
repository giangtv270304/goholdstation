package goholdstation

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// DepositVATransaction is one VND transfer received on a deposit account.
type DepositVATransaction struct {
	// ID is the fund log id.
	ID string `json:"id"`
	// Amount is the VND credited after Fee; both are decimal strings.
	Amount            string `json:"amount"`
	Fee               string `json:"fee"`
	SenderName        string `json:"sender_name"`
	SenderAccount     string `json:"sender_account"`
	SenderBankBin     string `json:"sender_bank_bin"`
	TransferContent   string `json:"transfer_content"`
	TransactionID     string `json:"transaction_id"`
	BankTransactionID string `json:"bank_transaction_id"`
	BankTraceID       string `json:"bank_trace_id"`
	// CreatedAt is an RFC 3339 string.
	CreatedAt string `json:"created_at"`
}

// ListDepositVATransactionsRequest: zero Page and PageSize are left out.
type ListDepositVATransactionsRequest struct {
	VAID string
	Page int
	// PageSize defaults to 20; the maximum is 100.
	PageSize int
}

type ListDepositVATransactionsResponse struct {
	Transactions []DepositVATransaction `json:"transactions"`
	Page         int                    `json:"page"`
	PageSize     int                    `json:"page_size"`
	TotalCount   int                    `json:"total_count"`
}

// ListDepositVATransactions returns the transfers received on one deposit account, newest first. Signed.
// It fails with ErrMissingField when VAID is empty.
func (c *Client) ListDepositVATransactions(ctx context.Context, req *ListDepositVATransactionsRequest) (*Response[ListDepositVATransactionsResponse], error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	if req.VAID == "" {
		return nil, fmt.Errorf("%w: VAID", ErrMissingField)
	}

	query := url.Values{}
	setPositive(query, "page", int64(req.Page))
	setPositive(query, "page_size", int64(req.PageSize))

	return call[ListDepositVATransactionsResponse](ctx, c, request{
		Method: http.MethodGet,
		Path:   "/partners/funds/deposit-vas/" + url.PathEscape(req.VAID) + "/transactions",
		Query:  query,
		Body:   nil,
		Signed: true,
	})
}
