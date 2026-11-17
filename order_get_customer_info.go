package goholdstation

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// Codes of GetOrderCustomerInfo. NotSupported and NotAvailable are states, not faults.
const (
	CustomerInfoCodeSuccess      = 0
	CustomerInfoCodeNotSupported = 2
	CustomerInfoCodeNotAvailable = 3
)

// CustomerInfo is the bank account that paid an onramp order. CustomerAccount is masked.
type CustomerInfo struct {
	CustomerAccount  string `json:"customer_account"`
	CustomerBankCode string `json:"customer_bank_code"`
	CustomerName     string `json:"customer_name"`
}

type GetOrderCustomerInfoRequest struct {
	OrderID string
}

// GetOrderCustomerInfoResponse: CustomerInfo is nil while Success is false (Code 2 or 3).
type GetOrderCustomerInfoResponse struct {
	BusinessResult

	CustomerInfo *CustomerInfo `json:"customer_info"`
}

// GetOrderCustomerInfo returns who paid an onramp order. Success=false is returned in the result, not as an error.
// It fails with ErrMissingField when OrderID is empty.
func (c *Client) GetOrderCustomerInfo(ctx context.Context, req *GetOrderCustomerInfoRequest) (*Response[GetOrderCustomerInfoResponse], error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	if req.OrderID == "" {
		return nil, fmt.Errorf("%w: OrderID", ErrMissingField)
	}

	return call[GetOrderCustomerInfoResponse](ctx, c, request{
		Method:               http.MethodGet,
		Path:                 "/partners/orders/" + url.PathEscape(req.OrderID) + "/customer-info",
		Query:                nil,
		Body:                 nil,
		AllowBusinessFailure: true,
	})
}
