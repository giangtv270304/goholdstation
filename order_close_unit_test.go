package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestCloseOrder_Success(t *testing.T) {
	t.Parallel()

	client, api := newFakeAPI(t, http.StatusOK, `{"message":"success","data":{"success":true,"code":0,"error":""}}`)

	resp, err := client.CloseOrder(context.Background(), &goholdstation.CloseOrderRequest{OrderID: "ord_1"})
	if err != nil {
		t.Fatalf("CloseOrder: %v", err)
	}

	api.assertRequest(t, wantRequest{method: http.MethodPost, path: "/partners/orders/ord_1/close", query: "", body: "", signed: false})

	if !resp.Data.Success || resp.Data.Code != 0 || resp.Data.Error != "" {
		t.Errorf("data = %+v", resp.Data)
	}
}

func TestCloseOrder_BusinessFailureIsAnAPIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusOK, `{"message":"success","data":{"success":false,"code":7,"error":"order is not pending"}}`)

	_, err := client.CloseOrder(context.Background(), &goholdstation.CloseOrderRequest{OrderID: "ord_1"})

	apiErr := requireAPIError(t, err, http.StatusOK, "7")
	if apiErr.Message != "order is not pending" {
		t.Errorf("message = %q", apiErr.Message)
	}
}

func TestCloseOrder_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusBadRequest, badRequestBody)

	_, err := client.CloseOrder(context.Background(), &goholdstation.CloseOrderRequest{OrderID: "ord_1"})
	assertAPIError(t, err, http.StatusBadRequest, "BAD_REQUEST")
}
