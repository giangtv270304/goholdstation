package goholdstation

import (
	"context"
	"encoding/json"
	"net/http"
)

// Config is one server-side configuration entry.
type Config struct {
	Key         string `json:"key"`
	Description string `json:"description"`
	// Value is kept raw: its shape is not documented. Amounts in it are in VND.
	Value     json.RawMessage `json:"value"`
	CreatedAt *Timestamp      `json:"created_at"`
	UpdatedAt *Timestamp      `json:"updated_at"`
}

type GetTransactionThresholdRequest struct{}

type GetTransactionThresholdResponse struct {
	Config *Config `json:"config"`
}

// GetTransactionThreshold returns the transaction threshold configuration. Pass nil.
func (c *Client) GetTransactionThreshold(ctx context.Context, _ *GetTransactionThresholdRequest) (*Response[GetTransactionThresholdResponse], error) {
	return call[GetTransactionThresholdResponse](ctx, c, request{
		Method: http.MethodGet,
		Path:   "/api/v1/configs/transaction_threshold",
		Query:  nil,
		Body:   nil,
	})
}
