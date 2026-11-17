package goholdstation

import (
	"context"
	"net/http"
	"net/url"
)

// ListFundLogsRequest: zero values are left out and the API applies its defaults.
type ListFundLogsRequest struct {
	Token      string
	ChangeType FundChangeType
	Page       int
	// PageSize defaults to 20; the maximum is 100.
	PageSize int
}

type ListFundLogsResponse struct {
	BusinessResult

	Logs       []FundLog `json:"logs"`
	TotalCount int       `json:"total_count"`
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
}

// ListFundLogs returns one page of fund logs. Signed. Pass nil for the first page with defaults.
func (c *Client) ListFundLogs(ctx context.Context, req *ListFundLogsRequest) (*Response[ListFundLogsResponse], error) {
	query := url.Values{}

	if req != nil {
		setNonEmpty(query, "token", req.Token)
		setPositive(query, "change_type", int64(req.ChangeType))
		setPositive(query, "page", int64(req.Page))
		setPositive(query, "page_size", int64(req.PageSize))
	}

	return call[ListFundLogsResponse](ctx, c, request{
		Method: http.MethodGet,
		Path:   "/partners/funds/logs",
		Query:  query,
		Body:   nil,
		Signed: true,
	})
}
