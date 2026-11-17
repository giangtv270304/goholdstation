package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestCreateCryptoPayoutOrder_Success(t *testing.T) {
	t.Parallel()

	client, api := newFakeAPI(t, http.StatusOK, `{"message":"success","data":{"success":true,"code":0,"error":"","order":{"id":"ord_c1","state":2,"processing_state":40,"amount":"2600000","currency":"vnd","outcome":"100.5","recipient":"0x1111111111111111111111111111111111111111"}}}`)

	resp, err := client.CreateCryptoPayoutOrder(context.Background(), &goholdstation.CreateCryptoPayoutOrderRequest{
		Amount: "2600000", Currency: "vnd", TokenAddress: "0x55d398326f99059ff775485246999027b3197955", ChainID: 56,
		Recipient: "0x1111111111111111111111111111111111111111", Callback: "https://yourapp.com/cb",
		ReferenceNo: "ref-c1", IdempotencyKey: "idem-c1",
	})
	if err != nil {
		t.Fatalf("CreateCryptoPayoutOrder: %v", err)
	}

	api.assertRequest(t, wantRequest{
		method: http.MethodPost, path: "/partners/orders/crypto-payout", query: "",
		body:   `{"amount":"2600000","currency":"vnd","token_address":"0x55d398326f99059ff775485246999027b3197955","chain_id":56,"recipient":"0x1111111111111111111111111111111111111111","callback":"https://yourapp.com/cb","reference_no":"ref-c1","idempotency_key":"idem-c1"}`,
		signed: true,
	})

	if got := api.header.Get("X-Idempotency-Key"); got != "" {
		t.Errorf("X-Idempotency-Key = %q, want the key in the body only", got)
	}

	order := resp.Data.Order
	if order == nil || order.ProcessingState != goholdstation.OrderProcessingStateCryptoPayoutWaitingForApproval || order.Outcome != "100.5" {
		t.Errorf("order = %+v", order)
	}
}

func TestCreateCryptoPayoutOrder_BusinessFailureIsAnAPIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusOK, `{"message":"success","data":{"success":false,"code":12,"error":"under maintenance"}}`)

	_, err := client.CreateCryptoPayoutOrder(context.Background(), &goholdstation.CreateCryptoPayoutOrderRequest{Amount: "1"})
	assertAPIError(t, err, http.StatusOK, "12")
}

func TestCreateCryptoPayoutOrder_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusUnauthorized, unauthorizedBody)

	_, err := client.CreateCryptoPayoutOrder(context.Background(), &goholdstation.CreateCryptoPayoutOrderRequest{Amount: "1"})
	assertAPIError(t, err, http.StatusUnauthorized, "UNAUTHORIZED")
}
