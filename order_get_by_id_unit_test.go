package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestGetOrderByID_Success(t *testing.T) {
	t.Parallel()

	client, api := newFakeAPI(t, http.StatusOK, `{"message":"success","data":{"order":`+depositOrderJSON+`,"transaction":{"tx_hash":"0xabc","status":1,"chain_id":56,"block_number":123,"nonce_used":7,"order_id":"3c1f6f8e-6d0a-4f0e-9d7a-0b6d2f1a9c11","error_message":"","metadata":{"retry_id":"r1"},"body":{"x":1},"created_at":{"seconds":1700000100,"nanos":0}},"payments":[{"id":"pay_1","order_id":"3c1f6f8e-6d0a-4f0e-9d7a-0b6d2f1a9c11","kind":1,"amount":"1000000","currency":"vnd","external_ref":"bank-1","received_at":{"seconds":1700000050,"nanos":0}}],"total_received":"1000000"}}`)

	resp, err := client.GetOrderByID(context.Background(), &goholdstation.GetOrderByIDRequest{OrderID: "3c1f6f8e/..?x"})
	if err != nil {
		t.Fatalf("GetOrderByID: %v", err)
	}

	api.assertRequest(t, wantRequest{method: http.MethodGet, path: "/partners/orders/3c1f6f8e%2F..%3Fx", query: "", body: "", signed: false})

	data := resp.Data
	if data.Order == nil || data.Order.Code != 100001 || data.Order.Fees == nil || data.Order.Fees.MinFee != "0.01" {
		t.Fatalf("order = %+v", data.Order)
	}

	transaction := data.Transaction
	if transaction == nil || transaction.TxHash != "0xabc" || transaction.BlockNumber != 123 || transaction.Metadata.RetryID != "r1" || string(transaction.Body) != `{"x":1}` {
		t.Errorf("transaction = %+v", transaction)
	}

	if len(data.Payments) != 1 || data.Payments[0].Amount != "1000000" || data.Payments[0].ReceivedAt.Seconds != 1700000050 || data.TotalReceived != "1000000" {
		t.Errorf("payments = %+v total = %q", data.Payments, data.TotalReceived)
	}
}

func TestGetOrderByID_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusNotFound, `{"error":"not found","error_code":"ORDER_NOT_FOUND","message":"order not found"}`)

	_, err := client.GetOrderByID(context.Background(), &goholdstation.GetOrderByIDRequest{OrderID: "missing"})
	if !goholdstation.IsNotFound(err) {
		t.Errorf("error = %v, want IsNotFound", err)
	}

	assertAPIError(t, err, http.StatusNotFound, "ORDER_NOT_FOUND")
}
