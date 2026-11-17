package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestCloseDepositVA_Success(t *testing.T) {
	t.Parallel()

	client, api := newFakeAPI(t, http.StatusOK, `{"message":"success","data":{"id":"dva_01HQ8X2M4K","account_number":"900112345678","account_name":"SDK SHOP","bank_id":17,"status":"CLOSED","created_at":"2026-10-29T03:14:22Z"}}`)

	resp, err := client.CloseDepositVA(context.Background(), &goholdstation.CloseDepositVARequest{VAID: "dva_01HQ8X2M4K"})
	if err != nil {
		t.Fatalf("CloseDepositVA: %v", err)
	}

	api.assertRequest(t, wantRequest{method: http.MethodPost, path: "/partners/funds/deposit-vas/dva_01HQ8X2M4K/close", query: "", body: "", signed: true})

	if resp.Data.ID != "dva_01HQ8X2M4K" || resp.Data.Status != goholdstation.DepositVAStatusClosed {
		t.Errorf("account = %+v", resp.Data)
	}
}

func TestCloseDepositVA_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusNotFound, `{"error":"not found","error_code":"NOT_FOUND","message":"No deposit VA with this ID"}`)

	_, err := client.CloseDepositVA(context.Background(), &goholdstation.CloseDepositVARequest{VAID: "dva_missing"})
	assertAPIError(t, err, http.StatusNotFound, "NOT_FOUND")
}
