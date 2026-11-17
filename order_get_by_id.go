package goholdstation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// OrderType is the kind of order. Values for prefund withdraw and crypto payout orders are not documented.
type OrderType int32

const (
	OrderTypeOnramp  OrderType = 1
	OrderTypeOfframp OrderType = 2
)

// OrderState is the main state of an order. Unknown values decode without error.
type OrderState int32

const (
	OrderStateCreated    OrderState = 1
	OrderStateProcessing OrderState = 2
	OrderStateCompleted  OrderState = 3
	OrderStateFailed     OrderState = 4
	// OrderStateClosed: expired or closed manually.
	OrderStateClosed OrderState = 5
	// OrderStateRefunded appears only in the vendor webhook example.
	OrderStateRefunded OrderState = 6
	// OrderStateRejected: rejected by an operator after manual review.
	OrderStateRejected OrderState = 8
)

// OrderProcessingState is the step inside OrderStateProcessing. Unknown values decode without error.
type OrderProcessingState int32

const (
	OrderProcessingStateOnrampFiatPending            OrderProcessingState = 10
	OrderProcessingStateOnrampFiatConfirmed          OrderProcessingState = 11
	OrderProcessingStateOnrampFiatFailed             OrderProcessingState = 12
	OrderProcessingStateOnrampCryptoPending          OrderProcessingState = 13
	OrderProcessingStateOnrampCryptoConfirmed        OrderProcessingState = 14
	OrderProcessingStateOnrampCryptoFailed           OrderProcessingState = 15
	OrderProcessingStateOnrampFiatWaitingForApproval OrderProcessingState = 16
	OrderProcessingStateOnrampNeedToReview           OrderProcessingState = 17
	OrderProcessingStateOnrampReviewRejected         OrderProcessingState = 18

	OrderProcessingStateOfframpCryptoPending   OrderProcessingState = 20
	OrderProcessingStateOfframpCryptoConfirmed OrderProcessingState = 21
	OrderProcessingStateOfframpCryptoFailed    OrderProcessingState = 22
	OrderProcessingStateOfframpFiatPending     OrderProcessingState = 23
	OrderProcessingStateOfframpFiatConfirmed   OrderProcessingState = 24
	OrderProcessingStateOfframpFiatFailed      OrderProcessingState = 25

	OrderProcessingStateWithdrawFiatQueued     OrderProcessingState = 31
	OrderProcessingStateWithdrawFiatProcessing OrderProcessingState = 32
	OrderProcessingStateWithdrawFiatConfirmed  OrderProcessingState = 33
	OrderProcessingStateWithdrawFiatFailed     OrderProcessingState = 34

	OrderProcessingStateCryptoPayoutWaitingForApproval OrderProcessingState = 40
	OrderProcessingStateCryptoPayoutCryptoPending      OrderProcessingState = 41
	OrderProcessingStateCryptoPayoutCryptoConfirmed    OrderProcessingState = 42
	// OrderProcessingStateCryptoPayoutCryptoFailed: the VND is refunded and the payout can be settled again.
	OrderProcessingStateCryptoPayoutCryptoFailed OrderProcessingState = 43
	OrderProcessingStateCryptoPayoutCryptoQueued OrderProcessingState = 44
)

// AppliedFee is the fee applied to an order. The unit of each value is not documented.
type AppliedFee struct {
	ServiceFee Decimal `json:"service_fee"`
	MinFee     Decimal `json:"min_fee"`
}

// PayData tells the payer how to pay an onramp order (bank account and QR code).
type PayData struct {
	BankAccountName   string `json:"bank_account_name"`
	BankAccountNumber string `json:"bank_account_number"`
	BankBinCode       string `json:"bank_bin_code"`
	BankName          string `json:"bank_name"`
	BankShortCode     string `json:"bank_short_code"`
	PaymentLink       string `json:"payment_link"`
	QRCode            string `json:"qr_code"`
	QRImage           string `json:"qr_image"`
	ExternalID        string `json:"external_id"`
}

// PaymentInfo is the bank account that receives a VND payout.
type PaymentInfo struct {
	// BankID is Bank.ID from ListBanks.
	BankID   int64  `json:"bank_id"`
	FullName string `json:"full_name"`
	// AccountType is 1 for a bank account number, 2 for an ATM card number.
	AccountType   int32  `json:"account_type"`
	AccountNumber string `json:"account_number"`
	// TransferNote is an optional memo, at most 100 characters.
	TransferNote string `json:"transfer_note,omitempty"`
}

// OrderError describes why an order failed.
type OrderError struct {
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
	Details      string `json:"details"`
	Retryable    bool   `json:"retryable"`
	Source       int32  `json:"source"`
	Timestamp    string `json:"timestamp"`
}

// OrderReference holds the rates used to price an order. Present only on orders from ListOrders.
type OrderReference struct {
	HSRate                 Decimal `json:"hs_rate"`
	USDTRate               Decimal `json:"usdt_rate"`
	TokenRate              Decimal `json:"token_rate"`
	PartnerRate            Decimal `json:"partner_rate"`
	Spread                 Decimal `json:"spread"`
	BufferRate             Decimal `json:"buffer_rate"`
	BufferedRateUSDT       Decimal `json:"buffered_rate_usdt"`
	BufferedRatePartner    Decimal `json:"buffered_rate_partner"`
	AdditionShareRate      Decimal `json:"addition_share_rate"`
	PartnerOverrideRevenue Decimal `json:"partner_override_revenue"`
	ProviderFee            Decimal `json:"provider_fee"`
	TokenSymbol            string  `json:"token_symbol"`
	SimWebhookID           string  `json:"sim_webhook_id"`
}

// Order is an onramp, offramp, prefund withdraw or crypto payout order.
// Amount, Outcome, Rate and OriginalRate are decimal strings: never parse them into a float.
// Any field may be absent.
type Order struct {
	// ID is a UUID.
	ID              string               `json:"id"`
	Code            int64                `json:"code"`
	OrderType       OrderType            `json:"order_type"`
	State           OrderState           `json:"state"`
	ProcessingState OrderProcessingState `json:"processing_state"`
	// Amount is in Currency ("vnd", "token" or "usd").
	Amount          string      `json:"amount"`
	Currency        string      `json:"currency"`
	Outcome         string      `json:"outcome"`
	OutcomeCurrency string      `json:"outcome_currency"`
	Rate            string      `json:"rate"`
	OriginalRate    string      `json:"original_rate"`
	Fees            *AppliedFee `json:"fees"`
	TotalFeeVND     Decimal     `json:"total_fee_vnd"`
	TokenAddress    string      `json:"token_address"`
	ChainID         int64       `json:"chain_id"`
	// Recipient is the user wallet for onramp, and the Holdstation wallet to pay for offramp.
	Recipient     string `json:"recipient"`
	RefundAddress string `json:"refund_address"`
	ReferenceNo   string `json:"reference_no"`
	ExternalID    string `json:"external_id"`
	PartnerID     string `json:"partner_id"`
	UserID        string `json:"user_id"`
	Provider      int32  `json:"provider"`
	// ProcessingTime has no documented unit.
	ProcessingTime int64        `json:"processing_time"`
	ClientIP       string       `json:"client_ip"`
	Callback       string       `json:"callback"`
	Purpose        string       `json:"purpose"`
	Note           string       `json:"note"`
	PayData        *PayData     `json:"pay_data"`
	PaymentInfo    *PaymentInfo `json:"payment_info"`
	Error          *OrderError  `json:"error"`
	CreatedAt      *Timestamp   `json:"created_at"`
	UpdatedAt      *Timestamp   `json:"updated_at"`
	ExpiredAt      *Timestamp   `json:"expired_at"`

	// KYCID, ApprovalReason, Body and Reference are returned by ListOrders only.
	KYCID          string          `json:"kyc_id"`
	ApprovalReason int32           `json:"approval_reason"`
	Body           json.RawMessage `json:"body"`
	Reference      *OrderReference `json:"reference"`
}

// OrderTransactionMetadata is the metadata of an on-chain transaction.
type OrderTransactionMetadata struct {
	RetryID string `json:"retry_id"`
}

// OrderTransaction is the on-chain transaction of an order.
type OrderTransaction struct {
	TxHash string `json:"tx_hash"`
	// Status values are not documented.
	Status       int32                     `json:"status"`
	ChainID      int64                     `json:"chain_id"`
	BlockNumber  int64                     `json:"block_number"`
	NonceUsed    int64                     `json:"nonce_used"`
	OrderID      string                    `json:"order_id"`
	ErrorMessage string                    `json:"error_message"`
	Metadata     *OrderTransactionMetadata `json:"metadata"`
	Body         json.RawMessage           `json:"body"`
	CreatedAt    *Timestamp                `json:"created_at"`
}

// OrderPayment is one payment received for an order. Amount is a decimal string in Currency.
type OrderPayment struct {
	ID      string `json:"id"`
	OrderID string `json:"order_id"`
	// Kind values are not documented.
	Kind        int32      `json:"kind"`
	Amount      string     `json:"amount"`
	Currency    string     `json:"currency"`
	ExternalRef string     `json:"external_ref"`
	ReceivedAt  *Timestamp `json:"received_at"`
}

type GetOrderByIDRequest struct {
	OrderID string
}

type GetOrderByIDResponse struct {
	Order       *Order            `json:"order"`
	Transaction *OrderTransaction `json:"transaction"`
	Payments    []OrderPayment    `json:"payments"`
	// TotalReceived is a decimal string.
	TotalReceived string `json:"total_received"`
}

// GetOrderByID returns one order. It fails with ErrMissingField when OrderID is empty.
func (c *Client) GetOrderByID(ctx context.Context, req *GetOrderByIDRequest) (*Response[GetOrderByIDResponse], error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	if req.OrderID == "" {
		return nil, fmt.Errorf("%w: OrderID", ErrMissingField)
	}

	return call[GetOrderByIDResponse](ctx, c, request{
		Method: http.MethodGet,
		Path:   "/partners/orders/" + url.PathEscape(req.OrderID),
		Query:  nil,
		Body:   nil,
	})
}
