package goholdstation

import (
	"context"
	"net/http"
)

type CreateMIDVARequest struct {
	// VAName is at most 200 characters, Remark at most 50.
	VAName string `json:"va_name"`
	Remark string `json:"remark"`
	// VAExpirationTime is Unix seconds; nil is sent as null and asks for the maximum lifetime.
	VAExpirationTime *int64 `json:"va_expiration_time"`
}

type CreateMIDVAResponse = MIDVA

// CreateMIDVA creates a multi-use merchant virtual account. Signed.
func (c *Client) CreateMIDVA(ctx context.Context, req *CreateMIDVARequest) (*Response[CreateMIDVAResponse], error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	return call[CreateMIDVAResponse](ctx, c, request{
		Method: http.MethodPost,
		Path:   "/partners/mid/va",
		Query:  nil,
		Body:   req,
		Signed: true,
	})
}
