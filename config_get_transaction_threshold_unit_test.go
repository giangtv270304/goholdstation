package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestGetTransactionThreshold_Success(t *testing.T) {
	t.Parallel()

	client, api := newFakeAPI(t, http.StatusOK, `{"message":"success","data":{"config":{"key":"transaction_threshold","description":"All values are in VND","value":{"fields":{}},"created_at":{"seconds":1700000000,"nanos":0},"updated_at":{"seconds":1700000001,"nanos":0}}}}`)

	resp, err := client.GetTransactionThreshold(context.Background(), &goholdstation.GetTransactionThresholdRequest{})
	if err != nil {
		t.Fatalf("GetTransactionThreshold: %v", err)
	}

	api.assertRequest(t, wantRequest{method: http.MethodGet, path: "/api/v1/configs/transaction_threshold", query: "", body: "", signed: false})

	config := resp.Data.Config
	if config == nil || config.Key != "transaction_threshold" || string(config.Value) != `{"fields":{}}` || config.UpdatedAt.Seconds != 1700000001 {
		t.Errorf("config = %+v", config)
	}
}

func TestGetTransactionThreshold_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusInternalServerError, `{"error":"internal","error_code":"INTERNAL","message":"boom"}`)

	_, err := client.GetTransactionThreshold(context.Background(), nil)
	assertAPIError(t, err, http.StatusInternalServerError, "INTERNAL")
}
