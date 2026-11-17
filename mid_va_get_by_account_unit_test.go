package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestGetMIDVAByAccount_Success(t *testing.T) {
	t.Parallel()

	client, api := newFakeAPI(t, http.StatusOK, `{"message":"success","data":{"id":"va_1","partner_id":"partner-xyz","va_account":"9704361234","va_name":"SDK SHOP","va_bank":"BIDV","remark":"sdk","status":1,"created_at":{"seconds":1793243662,"nanos":0},"updated_at":{"seconds":1793243662,"nanos":0}}}`)

	resp, err := client.GetMIDVAByAccount(context.Background(), &goholdstation.GetMIDVAByAccountRequest{VAAccount: "9704361234"})
	if err != nil {
		t.Fatalf("GetMIDVAByAccount: %v", err)
	}

	api.assertRequest(t, wantRequest{method: http.MethodGet, path: "/partners/mid/va/9704361234", query: "", body: "", signed: true})

	account := resp.Data
	if account.ID != "va_1" || account.PartnerID != "partner-xyz" || account.VAName != "SDK SHOP" || account.Remark != "sdk" || account.Status != goholdstation.MIDVAStatusActive {
		t.Errorf("account = %+v", account)
	}
}

func TestGetMIDVAByAccount_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusNotFound, notFoundBody)

	_, err := client.GetMIDVAByAccount(context.Background(), &goholdstation.GetMIDVAByAccountRequest{VAAccount: "0"})
	assertAPIError(t, err, http.StatusNotFound, "NOT_FOUND")
}
