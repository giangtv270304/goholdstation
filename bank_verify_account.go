package goholdstation

import (
	"context"
	"net/http"
)

type VerifyBankAccountRequest struct {
	BinCode       string `json:"bin_code"`
	AccountNumber string `json:"account_number"`
}

type VerifyBankAccountResponse struct {
	Bank          *Bank  `json:"bank"`
	AccountName   string `json:"account_name"`
	AccountNumber string `json:"account_number"`
	// AccountType is 1 for a bank account number, 2 for a card number.
	AccountType int32 `json:"account_type"`
}

// VerifyBankAccount looks up the holder name of a bank account.
func (c *Client) VerifyBankAccount(ctx context.Context, req *VerifyBankAccountRequest) (*Response[VerifyBankAccountResponse], error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	return call[VerifyBankAccountResponse](ctx, c, request{
		Method: http.MethodPost,
		Path:   "/api/v1/payment-methods/verify",
		Query:  nil,
		Body:   req,
	})
}
