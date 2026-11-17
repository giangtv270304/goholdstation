package goholdstation

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

// ListKYCsRequest: zero values are left out and the API applies its defaults.
type ListKYCsRequest struct {
	Page int
	// PageSize defaults to 20; the maximum is 100.
	PageSize int
}

// ListKYCsResponse: the paging fields are zero when the API answers with a plain array.
// TotalCount is also left out by the API when it is zero.
type ListKYCsResponse struct {
	KYCs       []KYC `json:"kycs"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalCount int   `json:"total_count"`
}

// UnmarshalJSON accepts the documented paged object and the documented plain array.
func (r *ListKYCsResponse) UnmarshalJSON(data []byte) error {
	if isJSONArray(data) {
		return json.Unmarshal(data, &r.KYCs)
	}

	type plain ListKYCsResponse

	return json.Unmarshal(data, (*plain)(r))
}

// ListKYCs returns one page of KYC records. Signed. Pass nil for the first page with defaults.
func (c *Client) ListKYCs(ctx context.Context, req *ListKYCsRequest) (*Response[ListKYCsResponse], error) {
	query := url.Values{}

	if req != nil {
		setPositive(query, "page", int64(req.Page))
		setPositive(query, "page_size", int64(req.PageSize))
	}

	return call[ListKYCsResponse](ctx, c, request{
		Method: http.MethodGet,
		Path:   "/partners/kycs",
		Query:  query,
		Body:   nil,
		Signed: true,
	})
}
