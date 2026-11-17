package goholdstation

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type CloseDepositVARequest struct {
	VAID string
}

type CloseDepositVAResponse = DepositVA

// CloseDepositVA closes a VND deposit account. Signed. Closing cannot be undone.
// It fails with ErrMissingField when VAID is empty.
func (c *Client) CloseDepositVA(ctx context.Context, req *CloseDepositVARequest) (*Response[CloseDepositVAResponse], error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	if req.VAID == "" {
		return nil, fmt.Errorf("%w: VAID", ErrMissingField)
	}

	return call[CloseDepositVAResponse](ctx, c, request{
		Method: http.MethodPost,
		Path:   "/partners/funds/deposit-vas/" + url.PathEscape(req.VAID) + "/close",
		Query:  nil,
		Body:   nil,
		Signed: true,
	})
}
