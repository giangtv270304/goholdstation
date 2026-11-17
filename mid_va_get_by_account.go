package goholdstation

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// MIDVAStatus is the state of a merchant virtual account. Unknown values decode without error.
type MIDVAStatus int32

const (
	MIDVAStatusActive MIDVAStatus = 1
	// MIDVAStatusExpired: past its expiration time, or revoked.
	MIDVAStatusExpired MIDVAStatus = 2
	MIDVAStatusRevoked MIDVAStatus = 4
)

// MIDVA is a multi-use merchant virtual account.
type MIDVA struct {
	ID        string      `json:"id"`
	PartnerID string      `json:"partner_id"`
	VAAccount string      `json:"va_account"`
	VAName    string      `json:"va_name"`
	VABank    string      `json:"va_bank"`
	Remark    string      `json:"remark"`
	Status    MIDVAStatus `json:"status"`
	CreatedAt *Timestamp  `json:"created_at"`
	UpdatedAt *Timestamp  `json:"updated_at"`
	ExpiredAt *Timestamp  `json:"expired_at"`
}

type GetMIDVAByAccountRequest struct {
	VAAccount string
}

type GetMIDVAByAccountResponse = MIDVA

// GetMIDVAByAccount returns one merchant virtual account. Signed.
// It fails with ErrMissingField when VAAccount is empty.
func (c *Client) GetMIDVAByAccount(ctx context.Context, req *GetMIDVAByAccountRequest) (*Response[GetMIDVAByAccountResponse], error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	if req.VAAccount == "" {
		return nil, fmt.Errorf("%w: VAAccount", ErrMissingField)
	}

	return call[GetMIDVAByAccountResponse](ctx, c, request{
		Method: http.MethodGet,
		Path:   "/partners/mid/va/" + url.PathEscape(req.VAAccount),
		Query:  nil,
		Body:   nil,
		Signed: true,
	})
}
