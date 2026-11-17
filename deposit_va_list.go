package goholdstation

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

// Values of ListDepositVAsRequest.Status. The filter is numeric while DepositVA.Status is a string.
const (
	DepositVAStatusFilterActive = 1
	DepositVAStatusFilterClosed = 2
)

// ListDepositVAsRequest: zero values are left out and the API applies its defaults.
type ListDepositVAsRequest struct {
	Status int
	Page   int
	// PageSize defaults to 20; the maximum is 100.
	PageSize int
}

// ListDepositVAsResponse: the paging fields are zero when the API answers with a plain array.
type ListDepositVAsResponse struct {
	DepositVAs []DepositVA `json:"deposit_vas"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalCount int         `json:"total_count"`
}

// UnmarshalJSON accepts the documented paged object and the documented plain array.
func (r *ListDepositVAsResponse) UnmarshalJSON(data []byte) error {
	if isJSONArray(data) {
		return json.Unmarshal(data, &r.DepositVAs)
	}

	type plain ListDepositVAsResponse

	return json.Unmarshal(data, (*plain)(r))
}

// ListDepositVAs returns the VND deposit accounts, newest first. Signed. Pass nil for defaults.
func (c *Client) ListDepositVAs(ctx context.Context, req *ListDepositVAsRequest) (*Response[ListDepositVAsResponse], error) {
	query := url.Values{}

	if req != nil {
		setPositive(query, "status", int64(req.Status))
		setPositive(query, "page", int64(req.Page))
		setPositive(query, "page_size", int64(req.PageSize))
	}

	return call[ListDepositVAsResponse](ctx, c, request{
		Method: http.MethodGet,
		Path:   "/partners/funds/deposit-vas",
		Query:  query,
		Body:   nil,
		Signed: true,
	})
}
