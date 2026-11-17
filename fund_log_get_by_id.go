package goholdstation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// FundChangeType says why a fund balance changed. Unknown values decode without error.
type FundChangeType int32

const (
	FundChangeTypeDeposit        FundChangeType = 1
	FundChangeTypeWithdraw       FundChangeType = 2
	FundChangeTypeCompensate     FundChangeType = 3
	FundChangeTypeConvertOut     FundChangeType = 4
	FundChangeTypeConvertIn      FundChangeType = 5
	FundChangeTypeAdjustDecrease FundChangeType = 6
)

// FundLog is one change of a prefunded balance. The four amounts are decimal strings.
type FundLog struct {
	// ID equals fund_log_id in the balance webhook.
	ID            int32          `json:"id"`
	PartnerFundID int32          `json:"partner_fund_id"`
	ChangeType    FundChangeType `json:"change_type"`
	// ChangeAmount is net and signed. Gross = ChangeAmount + Fee.
	ChangeAmount string `json:"change_amount"`
	Fee          string `json:"fee"`
	PreBalance   string `json:"pre_balance"`
	PostBalance  string `json:"post_balance"`
	// Metadata is kept raw; decode it into FundLogMetadata.
	Metadata json.RawMessage `json:"metadata"`
	CausedAt *Timestamp      `json:"caused_at"`
}

// FundLogMetadata is the documented plain-JSON form of FundLog.Metadata. Branch on Source, not on ChangeType.
// Only Source and ReferenceID are always present.
type FundLogMetadata struct {
	// Source is "fiat_deposit", "crypto_deposit", "withdraw", "crypto_payout", "compensate" or "convert".
	Source      string `json:"source"`
	ReferenceID string `json:"reference_id"`
	Reason      string `json:"reason"`

	DepositVAID       string `json:"deposit_va_id"`
	AccountNumber     string `json:"account_number"`
	TransactionID     string `json:"transaction_id"`
	BankTransactionID string `json:"bank_transaction_id"`
	BankTraceID       string `json:"bank_trace_id"`
	SenderName        string `json:"sender_name"`
	SenderAccount     string `json:"sender_account"`
	SenderBankBin     string `json:"sender_bank_bin"`
	TransferContent   string `json:"transfer_content"`
	// PaidAt is RFC 3339 and may be empty.
	PaidAt string `json:"paid_at"`

	ChainID int64  `json:"chain_id"`
	TxHash  string `json:"tx_hash"`

	OrderID string `json:"order_id"`

	ConvertRef string `json:"convert_ref"`
	FromToken  string `json:"from_token"`
	ToToken    string `json:"to_token"`
	// Rate is a decimal string, always "1".
	Rate string `json:"rate"`
}

type GetFundLogByIDRequest struct {
	FundLogID int32
}

type GetFundLogByIDResponse struct {
	BusinessResult

	Log *FundLog `json:"log"`
}

// GetFundLogByID returns one fund log. Signed. An unknown id is a 404 *APIError.
// It fails with ErrMissingField when FundLogID is not positive.
func (c *Client) GetFundLogByID(ctx context.Context, req *GetFundLogByIDRequest) (*Response[GetFundLogByIDResponse], error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	if req.FundLogID <= 0 {
		return nil, fmt.Errorf("%w: FundLogID", ErrMissingField)
	}

	return call[GetFundLogByIDResponse](ctx, c, request{
		Method: http.MethodGet,
		Path:   "/partners/funds/logs/" + strconv.FormatInt(int64(req.FundLogID), 10),
		Query:  nil,
		Body:   nil,
		Signed: true,
	})
}
