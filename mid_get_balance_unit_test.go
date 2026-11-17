package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestGetMIDBalance_Success_BothDocumentedForms(t *testing.T) {
	t.Parallel()

	t.Run("openapi form, no body", func(t *testing.T) {
		t.Parallel()

		client, api := newFakeAPI(t, http.StatusOK, `{"message":"success","data":{"id":3,"partner_id":"partner-xyz","token":"VND","balance":"1000000.50","fund_type":2,"updated_at":{"seconds":1700000000,"nanos":0}}}`)

		resp, err := client.GetMIDBalance(context.Background(), nil)
		if err != nil {
			t.Fatalf("GetMIDBalance: %v", err)
		}

		api.assertRequest(t, wantRequest{method: http.MethodPost, path: "/partners/mid/balance", query: "", body: "", signed: true})

		data := resp.Data
		if data.ID != 3 || data.PartnerID != "partner-xyz" || data.Token != "VND" || data.Balance != "1000000.50" || data.FundType != goholdstation.FundTypeFiat || data.UpdatedAt.Seconds != 1700000000 {
			t.Errorf("data = %+v", data)
		}
	})

	for name, body := range map[string]string{
		"md form, string merchant id": `{"merchant_id":"M001","merchant_email":"ops@example.com","balance":"1,000,000"}`,
		"md form, number merchant id": `{"merchant_id":1001,"merchant_email":"ops@example.com","balance":"1,000,000"}`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client, api := newFakeAPI(t, http.StatusOK, body)

			resp, err := client.GetMIDBalance(context.Background(), &goholdstation.GetMIDBalanceRequest{RequestID: "unique-request-id"})
			if err != nil {
				t.Fatalf("GetMIDBalance: %v", err)
			}

			api.assertRequest(t, wantRequest{method: http.MethodPost, path: "/partners/mid/balance", query: "", body: `{"request_id":"unique-request-id"}`, signed: true})

			data := resp.Data
			if (data.MerchantID != "M001" && data.MerchantID != "1001") || data.MerchantEmail != "ops@example.com" || data.Balance != "1,000,000" {
				t.Errorf("data = %+v", data)
			}
		})
	}
}

func TestGetMIDBalance_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusUnauthorized, unauthorizedBody)

	_, err := client.GetMIDBalance(context.Background(), nil)
	assertAPIError(t, err, http.StatusUnauthorized, "UNAUTHORIZED")
}
