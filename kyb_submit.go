package goholdstation

import (
	"context"
	"net/http"
)

type SubmitKYBRequest = KYBSubmission

type SubmitKYBResponse struct {
	KYB *KYB `json:"kyb"`
}

// SubmitKYB submits a business for verification. Signed. The upload can take up to 120 seconds:
// use an HTTP client with a longer timeout than DefaultTimeout.
// A Pending or Approved record already existing is a 409 *APIError; validation failures are in APIError.Details.
func (c *Client) SubmitKYB(ctx context.Context, req *SubmitKYBRequest) (*Response[SubmitKYBResponse], error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	return call[SubmitKYBResponse](ctx, c, request{
		Method: http.MethodPost,
		Path:   "/partners/kyb",
		Query:  nil,
		Body:   req,
		Signed: true,
	})
}
