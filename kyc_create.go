package goholdstation

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type CreateKYCRequest struct {
	// ID is chosen by the partner and becomes the kyc_id used on orders.
	ID  string
	KYC KYCInput
}

type CreateKYCResponse = KYCResult

// CreateKYC stores a customer identity record under the partner's id. Signed. The upload can take up to
// 120 seconds. An id that exists is a 409 *APIError with code "KYC_ALREADY_EXISTS". Records cannot be deleted.
// It fails with ErrMissingField when ID is empty.
func (c *Client) CreateKYC(ctx context.Context, req *CreateKYCRequest) (*Response[CreateKYCResponse], error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	if req.ID == "" {
		return nil, fmt.Errorf("%w: ID", ErrMissingField)
	}

	return call[CreateKYCResponse](ctx, c, request{
		Method: http.MethodPost,
		Path:   "/partners/kycs/" + url.PathEscape(req.ID),
		Query:  nil,
		Body:   req.KYC,
		Signed: true,
	})
}
