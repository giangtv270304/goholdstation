package goholdstation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// MIDVATransaction is one payment received on a merchant virtual account.
type MIDVATransaction struct {
	TransID int64 `json:"trans_id"`
	// Amount is in VND.
	Amount Decimal `json:"amount"`
	// TimePaid is Unix seconds.
	TimePaid int64  `json:"time_paid"`
	CashinID int64  `json:"cashin_id"`
	Remark   string `json:"remark"`
}

// UnmarshalJSON accepts transId/timePaid/cashinId and trans_id/time_paid/cashin_id.
func (m *MIDVATransaction) UnmarshalJSON(data []byte) error {
	var shadow struct {
		TransIDCamel  *int64  `json:"transId"`
		TransIDSnake  *int64  `json:"trans_id"`
		Amount        Decimal `json:"amount"`
		TimePaidCamel *int64  `json:"timePaid"`
		TimePaidSnake *int64  `json:"time_paid"`
		CashinIDCamel *int64  `json:"cashinId"`
		CashinIDSnake *int64  `json:"cashin_id"`
		Remark        string  `json:"remark"`
	}

	if err := json.Unmarshal(data, &shadow); err != nil {
		return err
	}

	*m = MIDVATransaction{
		TransID:  firstInt64(shadow.TransIDCamel, shadow.TransIDSnake),
		Amount:   shadow.Amount,
		TimePaid: firstInt64(shadow.TimePaidCamel, shadow.TimePaidSnake),
		CashinID: firstInt64(shadow.CashinIDCamel, shadow.CashinIDSnake),
		Remark:   shadow.Remark,
	}

	return nil
}

func firstInt64(candidates ...*int64) int64 {
	for _, candidate := range candidates {
		if candidate != nil {
			return *candidate
		}
	}

	return 0
}

type InquireMIDVARequest struct {
	VAAccount string
}

// InquireMIDVAResponse: the API documents camelCase and snake_case keys; both decode into these fields.
type InquireMIDVAResponse struct {
	MerchantID   string             `json:"merchant_id"`
	VAAccount    string             `json:"va_account"`
	VAName       string             `json:"va_name"`
	VABank       string             `json:"va_bank"`
	Transactions []MIDVATransaction `json:"transactions"`
}

// UnmarshalJSON accepts merchantId/vaAccount/vaName/vaBank and their snake_case forms.
func (r *InquireMIDVAResponse) UnmarshalJSON(data []byte) error {
	var shadow struct {
		MerchantIDCamel json.RawMessage    `json:"merchantId"`
		MerchantIDSnake json.RawMessage    `json:"merchant_id"`
		VAAccountCamel  json.RawMessage    `json:"vaAccount"`
		VAAccountSnake  json.RawMessage    `json:"va_account"`
		VANameCamel     json.RawMessage    `json:"vaName"`
		VANameSnake     json.RawMessage    `json:"va_name"`
		VABankCamel     json.RawMessage    `json:"vaBank"`
		VABankSnake     json.RawMessage    `json:"va_bank"`
		Transactions    []MIDVATransaction `json:"transactions"`
	}

	if err := json.Unmarshal(data, &shadow); err != nil {
		return err
	}

	*r = InquireMIDVAResponse{
		MerchantID:   rawToString(firstRaw(shadow.MerchantIDCamel, shadow.MerchantIDSnake)),
		VAAccount:    rawToString(firstRaw(shadow.VAAccountCamel, shadow.VAAccountSnake)),
		VAName:       rawToString(firstRaw(shadow.VANameCamel, shadow.VANameSnake)),
		VABank:       rawToString(firstRaw(shadow.VABankCamel, shadow.VABankSnake)),
		Transactions: shadow.Transactions,
	}

	return nil
}

// InquireMIDVA asks the payment provider, live, for the payments received on a merchant virtual account. Signed.
// It fails with ErrMissingField when VAAccount is empty.
func (c *Client) InquireMIDVA(ctx context.Context, req *InquireMIDVARequest) (*Response[InquireMIDVAResponse], error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	if req.VAAccount == "" {
		return nil, fmt.Errorf("%w: VAAccount", ErrMissingField)
	}

	return call[InquireMIDVAResponse](ctx, c, request{
		Method: http.MethodGet,
		Path:   "/partners/mid/va/" + url.PathEscape(req.VAAccount) + "/inquiry",
		Query:  nil,
		Body:   nil,
		Signed: true,
	})
}
