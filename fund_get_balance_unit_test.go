package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestGetFundBalance_Success_BareAndWrapped(t *testing.T) {
	t.Parallel()

	const payload = `{"success":true,"code":0,"error":"","fund":{"id":1,"partner_id":"partner-xyz","token":"USDT","balance":"123456789.123456789012345678","fund_type":1,"updated_at":{"seconds":1700000000,"nanos":0}}}`

	for name, body := range map[string]string{"bare": payload, "wrapped": `{"message":"success","data":` + payload + `}`} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client, api := newFakeAPI(t, http.StatusOK, body)

			resp, err := client.GetFundBalance(context.Background(), &goholdstation.GetFundBalanceRequest{Token: "USDT"})
			if err != nil {
				t.Fatalf("GetFundBalance: %v", err)
			}

			api.assertRequest(t, wantRequest{method: http.MethodGet, path: "/partners/funds/balance", query: "token=USDT", body: "", signed: true})

			fund := resp.Data.Fund
			if fund == nil || fund.Token != "USDT" || fund.Balance != "123456789.123456789012345678" || fund.FundType != goholdstation.FundTypeCrypto {
				t.Errorf("fund = %+v", fund)
			}
		})
	}
}

func TestGetFundBalance_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusBadRequest, `{"error":"bad request","error_code":"INVALID_TOKEN","message":"unknown token"}`)

	_, err := client.GetFundBalance(context.Background(), &goholdstation.GetFundBalanceRequest{Token: "DOGE"})
	assertAPIError(t, err, http.StatusBadRequest, "INVALID_TOKEN")
}
