package goholdstation

import (
	"context"
	"net/http"
)

type CreateDepositVARequest struct {
	// VAName is the name the sender's bank shows, 1 to 50 characters.
	// The vendor documents it both as required and as "no request body": when empty, no body is sent.
	VAName string `json:"va_name"`
}

type CreateDepositVAResponse = DepositVA

// CreateDepositVA opens a new VND deposit account. Signed.
// It is not idempotent: every call opens a new account, so never retry it blindly.
func (c *Client) CreateDepositVA(ctx context.Context, req *CreateDepositVARequest) (*Response[CreateDepositVAResponse], error) {
	var body any
	if req != nil && req.VAName != "" {
		body = req
	}

	return call[CreateDepositVAResponse](ctx, c, request{
		Method: http.MethodPost,
		Path:   "/partners/funds/deposit-vas",
		Query:  nil,
		Body:   body,
		Signed: true,
	})
}
