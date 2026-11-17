package goholdstation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type RevokeMIDVARequest struct {
	VAAccount string
}

// RevokeMIDVAResponse: VAStatus is "4" once revoked.
// The API documents camelCase and snake_case keys; both decode into these fields.
type RevokeMIDVAResponse struct {
	MerchantID string `json:"merchant_id"`
	VAAccount  string `json:"va_account"`
	VAStatus   string `json:"va_status"`
}

// UnmarshalJSON accepts merchantId/vaAccount/vaStatus and merchant_id/va_account/va_status.
func (r *RevokeMIDVAResponse) UnmarshalJSON(data []byte) error {
	var shadow struct {
		MerchantIDCamel json.RawMessage `json:"merchantId"`
		MerchantIDSnake json.RawMessage `json:"merchant_id"`
		VAAccountCamel  json.RawMessage `json:"vaAccount"`
		VAAccountSnake  json.RawMessage `json:"va_account"`
		VAStatusCamel   json.RawMessage `json:"vaStatus"`
		VAStatusSnake   json.RawMessage `json:"va_status"`
	}

	if err := json.Unmarshal(data, &shadow); err != nil {
		return err
	}

	r.MerchantID = rawToString(firstRaw(shadow.MerchantIDCamel, shadow.MerchantIDSnake))
	r.VAAccount = rawToString(firstRaw(shadow.VAAccountCamel, shadow.VAAccountSnake))
	r.VAStatus = rawToString(firstRaw(shadow.VAStatusCamel, shadow.VAStatusSnake))

	return nil
}

// RevokeMIDVA revokes a merchant virtual account. Signed. It cannot be undone.
// It fails with ErrMissingField when VAAccount is empty.
func (c *Client) RevokeMIDVA(ctx context.Context, req *RevokeMIDVARequest) (*Response[RevokeMIDVAResponse], error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	if req.VAAccount == "" {
		return nil, fmt.Errorf("%w: VAAccount", ErrMissingField)
	}

	return call[RevokeMIDVAResponse](ctx, c, request{
		Method: http.MethodDelete,
		Path:   "/partners/mid/va/" + url.PathEscape(req.VAAccount),
		Query:  nil,
		Body:   nil,
		Signed: true,
	})
}
