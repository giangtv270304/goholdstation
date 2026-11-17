package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestGetExchangeRate_Success(t *testing.T) {
	t.Parallel()

	client, api := newFakeAPI(t, http.StatusOK, `{"message":"success","data":{"rate":{"buy":"25850","sell":"25750"}}}`)

	resp, err := client.GetExchangeRate(context.Background(), &goholdstation.GetExchangeRateRequest{Original: true})
	if err != nil {
		t.Fatalf("GetExchangeRate: %v", err)
	}

	api.assertRequest(t, wantRequest{method: http.MethodGet, path: "/api/v1/exchange/rate", query: "original=true", body: "", signed: false})

	if resp.Message != "success" || resp.Data.Rate == nil || resp.Data.Rate.Buy != "25850" || resp.Data.Rate.Sell != "25750" {
		t.Errorf("resp = %+v", resp)
	}
}

func TestGetExchangeRate_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusUnauthorized, unauthorizedBody)

	_, err := client.GetExchangeRate(context.Background(), nil)
	if !goholdstation.IsUnauthorized(err) {
		t.Errorf("error = %v, want IsUnauthorized", err)
	}

	assertAPIError(t, err, http.StatusUnauthorized, "UNAUTHORIZED")
}
