package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestListFunds_Success_BareAndWrapped(t *testing.T) {
	t.Parallel()

	const payload = `{"success":true,"code":0,"error":"","funds":[{"id":1,"partner_id":"partner-xyz","token":"USDT","balance":"1000.50","updated_at":{"seconds":1700000000,"nanos":0}}]}`

	for name, body := range map[string]string{"bare": payload, "wrapped": `{"message":"success","data":` + payload + `}`} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client, api := newFakeAPI(t, http.StatusOK, body)

			resp, err := client.ListFunds(context.Background(), &goholdstation.ListFundsRequest{})
			if err != nil {
				t.Fatalf("ListFunds: %v", err)
			}

			api.assertRequest(t, wantRequest{method: http.MethodGet, path: "/partners/funds", query: "", body: "", signed: true})

			if !resp.Data.Success || len(resp.Data.Funds) != 1 {
				t.Fatalf("data = %+v", resp.Data)
			}

			fund := resp.Data.Funds[0]
			if fund.ID != 1 || fund.PartnerID != "partner-xyz" || fund.Token != "USDT" || fund.Balance != "1000.50" || fund.UpdatedAt.Seconds != 1700000000 {
				t.Errorf("fund = %+v", fund)
			}

			if wantMessage := map[string]string{"bare": "", "wrapped": "success"}[name]; resp.Message != wantMessage {
				t.Errorf("message = %q, want %q", resp.Message, wantMessage)
			}
		})
	}
}

func TestListFunds_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusUnauthorized, unauthorizedBody)

	_, err := client.ListFunds(context.Background(), nil)
	assertAPIError(t, err, http.StatusUnauthorized, "UNAUTHORIZED")
}
