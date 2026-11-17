package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestGetDepositVAByID_Success(t *testing.T) {
	t.Parallel()

	client, api := newFakeAPI(t, http.StatusOK, `{"message":"success","data":`+depositVAJSON+`}`)

	resp, err := client.GetDepositVAByID(context.Background(), &goholdstation.GetDepositVAByIDRequest{VAID: "dva_01HQ8X2M4K"})
	if err != nil {
		t.Fatalf("GetDepositVAByID: %v", err)
	}

	api.assertRequest(t, wantRequest{method: http.MethodGet, path: "/partners/funds/deposit-vas/dva_01HQ8X2M4K", query: "", body: "", signed: true})

	if resp.Data.ID != "dva_01HQ8X2M4K" || resp.Data.AccountNumber != "900112345678" || resp.Data.Status != "ACTIVE" {
		t.Errorf("account = %+v", resp.Data)
	}
}

func TestGetDepositVAByID_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusNotFound, `{"error":"not found","error_code":"NOT_FOUND","message":"No deposit VA with this ID"}`)

	_, err := client.GetDepositVAByID(context.Background(), &goholdstation.GetDepositVAByIDRequest{VAID: "dva_missing"})
	assertAPIError(t, err, http.StatusNotFound, "NOT_FOUND")
}
