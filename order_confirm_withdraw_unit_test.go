package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestConfirmWithdrawOrder_Success(t *testing.T) {
	t.Parallel()

	client, api := newFakeAPI(t, http.StatusOK, `{"message":"success","data":{"success":true,"code":0,"error":"","order":{"id":"ord_p1","state":2,"processing_state":31,"amount":"2575000"}}}`)

	resp, err := client.ConfirmWithdrawOrder(context.Background(), &goholdstation.ConfirmWithdrawOrderRequest{OrderID: "ord_p1"})
	if err != nil {
		t.Fatalf("ConfirmWithdrawOrder: %v", err)
	}

	api.assertRequest(t, wantRequest{method: http.MethodPost, path: "/partners/orders/ord_p1/confirm-withdraw", query: "", body: "", signed: true})

	if got := api.header.Get("Content-Type"); got != "" {
		t.Errorf("Content-Type = %q, want none for a request without body", got)
	}

	order := resp.Data.Order
	if order == nil || order.State != goholdstation.OrderStateProcessing || order.ProcessingState != goholdstation.OrderProcessingStateWithdrawFiatQueued || order.Amount != "2575000" {
		t.Errorf("order = %+v", order)
	}
}

func TestConfirmWithdrawOrder_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusNotFound, `{"error":"not found","error_code":"ORDER_EXPIRED","message":"order expired"}`)

	_, err := client.ConfirmWithdrawOrder(context.Background(), &goholdstation.ConfirmWithdrawOrderRequest{OrderID: "ord_p1"})
	assertAPIError(t, err, http.StatusNotFound, "ORDER_EXPIRED")
}
