package goholdstation

import (
	"context"
	"net/http"
)

// Business codes of CreateWithdrawOrder and ConfirmWithdrawOrder, found in APIError.Code as decimal strings.
const (
	WithdrawCodeInvalidInput           = 1
	WithdrawCodePartnerNotFound        = 2
	WithdrawCodeInsufficientBalance    = 3
	WithdrawCodeDuplicateRequest       = 4
	WithdrawCodeOrderNotFound          = 5
	WithdrawCodeOrderNotOwnedByPartner = 6
	WithdrawCodeInvalidOrderState      = 7
	WithdrawCodeInvalidOrderType       = 8
	WithdrawCodeOrderTypeNotAvailable  = 9
	WithdrawCodeUnderMaintenance       = 11
	// WithdrawCodeInternalError is safe to retry after a short backoff.
	WithdrawCodeInternalError = 99
)

type CreateWithdrawOrderRequest struct {
	// Amount is a decimal string in Currency ("vnd" or "token").
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
	// TokenSymbol is the prefunded token to debit: "USDT", "USDC" or "USD1".
	TokenSymbol string       `json:"token_symbol"`
	PaymentInfo *PaymentInfo `json:"payment_info"`
	// Callback is the HTTPS URL that receives the order status webhook.
	Callback    string `json:"callback"`
	ReferenceNo string `json:"reference_no,omitempty"`
	// IdempotencyKey: a key already used is refused with code 4.
	IdempotencyKey string `json:"idempotency_key"`
	KYCID          string `json:"kyc_id,omitempty"`
	// Purpose is stored immutably.
	Purpose string `json:"purpose,omitempty"`
}

type CreateWithdrawOrderResponse struct {
	BusinessResult

	Order *Order `json:"order"`
}

// CreateWithdrawOrder is step 1 of a prefund withdrawal: it returns a quote held for up to 15 minutes
// and moves no money. Signed. A business failure is an *APIError with status 200 and the code in Code.
// Any error other than an *APIError with a 4xx status or a business code means the outcome is unknown:
// read the order with GetOrderByReferenceNo, or resend with the same IdempotencyKey and ReferenceNo, never new ones.
func (c *Client) CreateWithdrawOrder(ctx context.Context, req *CreateWithdrawOrderRequest) (*Response[CreateWithdrawOrderResponse], error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	return call[CreateWithdrawOrderResponse](ctx, c, request{
		Method: http.MethodPost,
		Path:   "/partners/orders/withdraw",
		Query:  nil,
		Body:   req,
		Signed: true,
	})
}
