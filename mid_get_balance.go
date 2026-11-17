package goholdstation

import (
	"context"
	"encoding/json"
	"net/http"
)

type GetMIDBalanceRequest struct {
	// RequestID is sent as the body {"request_id": ...} when set. One vendor page documents it, another documents no body.
	RequestID string `json:"request_id"`
}

// GetMIDBalanceResponse holds the fields of both documented forms; the fields of the other form stay empty.
// Balance is a decimal string; one vendor example shows thousands separators ("1,000,000").
type GetMIDBalanceResponse struct {
	ID        int64      `json:"id"`
	PartnerID string     `json:"partner_id"`
	Token     string     `json:"token"`
	Balance   string     `json:"balance"`
	FundType  int32      `json:"fund_type"`
	UpdatedAt *Timestamp `json:"updated_at"`

	MerchantID    string `json:"merchant_id"`
	MerchantEmail string `json:"merchant_email"`
}

// UnmarshalJSON accepts merchant_id as a JSON string or number.
func (r *GetMIDBalanceResponse) UnmarshalJSON(data []byte) error {
	type plain GetMIDBalanceResponse

	shadow := struct {
		*plain

		MerchantID json.RawMessage `json:"merchant_id"`
	}{plain: (*plain)(r), MerchantID: nil}

	if err := json.Unmarshal(data, &shadow); err != nil {
		return err
	}

	r.MerchantID = rawToString(shadow.MerchantID)

	return nil
}

// GetMIDBalance returns the merchant (MID) balance. Signed. Pass nil to send no body.
func (c *Client) GetMIDBalance(ctx context.Context, req *GetMIDBalanceRequest) (*Response[GetMIDBalanceResponse], error) {
	var body any
	if req != nil && req.RequestID != "" {
		body = req
	}

	return call[GetMIDBalanceResponse](ctx, c, request{
		Method: http.MethodPost,
		Path:   "/partners/mid/balance",
		Query:  nil,
		Body:   body,
		Signed: true,
	})
}
