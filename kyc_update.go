package goholdstation

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type UpdateKYCRequest struct {
	ID  string
	KYC KYCInput
}

type UpdateKYCResponse = KYCResult

// UpdateKYC replaces the fields of a KYC record. Signed. An unknown id is a 404 *APIError with code "KYC_NOT_FOUND".
// It fails with ErrMissingField when ID is empty.
func (c *Client) UpdateKYC(ctx context.Context, req *UpdateKYCRequest) (*Response[UpdateKYCResponse], error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	if req.ID == "" {
		return nil, fmt.Errorf("%w: ID", ErrMissingField)
	}

	return call[UpdateKYCResponse](ctx, c, request{
		Method: http.MethodPut,
		Path:   "/partners/kycs/" + url.PathEscape(req.ID),
		Query:  nil,
		Body:   req.KYC,
		Signed: true,
	})
}
