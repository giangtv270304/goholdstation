package goholdstation

import (
	"context"
	"net/http"
)

// WholesaleApplication is a request for a wholesale rate.
type WholesaleApplication struct {
	ID      string `json:"id"`
	UserID  string `json:"user_id"`
	Email   string `json:"email"`
	Purpose string `json:"purpose"`
	// DefaultRate is a fraction: "0.01" is 1 percent.
	DefaultRate Decimal `json:"default_rate"`
	// Status values are not documented.
	Status      int32      `json:"status"`
	SubmittedAt *Timestamp `json:"submitted_at"`
	ReviewedAt  *Timestamp `json:"reviewed_at"`
	UpdatedAt   *Timestamp `json:"updated_at"`
}

type CreateWholesaleApplicationRequest struct {
	// DefaultRate is a fraction ("0.01" is 1 percent); nil is sent as null.
	DefaultRate *Decimal `json:"default_rate"`
}

type CreateWholesaleApplicationResponse struct {
	Application *WholesaleApplication `json:"application"`
}

// CreateWholesaleApplication submits a wholesale application.
// The vendor documents it "for the authenticated user": a partner key may be refused.
func (c *Client) CreateWholesaleApplication(ctx context.Context, req *CreateWholesaleApplicationRequest) (*Response[CreateWholesaleApplicationResponse], error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	return call[CreateWholesaleApplicationResponse](ctx, c, request{
		Method: http.MethodPost,
		Path:   "/api/v1/wholesale/applications",
		Query:  nil,
		Body:   req,
	})
}
