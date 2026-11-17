package goholdstation

import (
	"context"
	"net/http"
	"net/url"
)

// ListKYBsRequest: zero values are left out and the API applies its defaults.
type ListKYBsRequest struct {
	// ReferenceID is an exact match.
	ReferenceID string
	SubjectType int
	Status      KYBStatus
	Page        int
	// PageSize defaults to 20; the maximum is 100.
	PageSize int
}

type ListKYBsResponse struct {
	KYBs       []KYB `json:"kybs"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalCount int   `json:"total_count"`
}

// ListKYBs returns one page of KYB records, newest first. Signed. Pass nil for the first page with defaults.
func (c *Client) ListKYBs(ctx context.Context, req *ListKYBsRequest) (*Response[ListKYBsResponse], error) {
	query := url.Values{}

	if req != nil {
		setNonEmpty(query, "reference_id", req.ReferenceID)
		setPositive(query, "subject_type", int64(req.SubjectType))
		setPositive(query, "status", int64(req.Status))
		setPositive(query, "page", int64(req.Page))
		setPositive(query, "page_size", int64(req.PageSize))
	}

	return call[ListKYBsResponse](ctx, c, request{
		Method: http.MethodGet,
		Path:   "/partners/kyb",
		Query:  query,
		Body:   nil,
		Signed: true,
	})
}
