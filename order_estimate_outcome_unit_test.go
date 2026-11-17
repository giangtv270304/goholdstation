package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestEstimateOrderOutcome_Success(t *testing.T) {
	t.Parallel()

	client, api := newFakeAPI(t, http.StatusOK, `{"message":"success","data":{"outcome":"38.610038610038610038","rate":25900.123456789012345678,"fee_vnd":5000,"fees":{"service_fee":5000,"min_fee":0.01}}}`)

	resp, err := client.EstimateOrderOutcome(context.Background(), &goholdstation.EstimateOrderOutcomeRequest{
		Amount: "1000000", OrderType: goholdstation.OrderTypeOnramp, ChainID: 56,
		TokenAddress: "0x55d398326f99059ff775485246999027b3197955", Reverse: false, Buffer: decimalPtr("0.005"),
	})
	if err != nil {
		t.Fatalf("EstimateOrderOutcome: %v", err)
	}

	api.assertRequest(t, wantRequest{
		method: http.MethodPost, path: "/api/v1/payments/estimate-outcome", query: "",
		body:   `{"amount":"1000000","order_type":1,"chain_id":56,"token_address":"0x55d398326f99059ff775485246999027b3197955","reverse":false,"buffer":0.005}`,
		signed: false,
	})

	data := resp.Data
	if data.Outcome != "38.610038610038610038" || data.Rate != "25900.123456789012345678" || data.FeeVND != "5000" {
		t.Errorf("data = %+v, want every digit kept", data)
	}

	if data.Fees == nil || data.Fees.ServiceFee != "5000" || data.Fees.MinFee != "0.01" {
		t.Errorf("fees = %+v", data.Fees)
	}
}

func TestEstimateOrderOutcome_NilBufferIsLeftOut(t *testing.T) {
	t.Parallel()

	client, api := newFakeAPI(t, http.StatusOK, `{"message":"success","data":{"outcome":"1","rate":1,"fee_vnd":0}}`)

	_, err := client.EstimateOrderOutcome(context.Background(), &goholdstation.EstimateOrderOutcomeRequest{
		Amount: "10", OrderType: goholdstation.OrderTypeOfframp, ChainID: 56, TokenAddress: "0xabc", Reverse: true,
	})
	if err != nil {
		t.Fatalf("EstimateOrderOutcome: %v", err)
	}

	api.assertRequest(t, wantRequest{
		method: http.MethodPost, path: "/api/v1/payments/estimate-outcome", query: "",
		body: `{"amount":"10","order_type":2,"chain_id":56,"token_address":"0xabc","reverse":true}`, signed: false,
	})
}

func TestEstimateOrderOutcome_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusBadRequest, `{"error":"bad request","error_code":"INVALID_TOKEN","message":"token is not supported"}`)

	_, err := client.EstimateOrderOutcome(context.Background(), &goholdstation.EstimateOrderOutcomeRequest{Amount: "1"})
	assertAPIError(t, err, http.StatusBadRequest, "INVALID_TOKEN")
}
