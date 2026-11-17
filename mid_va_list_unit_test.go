package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestListMIDVAs_Success(t *testing.T) {
	t.Parallel()

	client, api := newFakeAPI(t, http.StatusOK, `{"message":"success","data":{"vas":[{"id":"va_1","partner_id":"partner-xyz","va_account":"9704361234","va_name":"SDK SHOP","va_bank":"BIDV","remark":"sdk","status":2,"created_at":"2026-10-29T03:14:22Z","updated_at":{"seconds":1793243662,"nanos":0}}],"total_count":1,"page":1,"page_size":20}}`)

	resp, err := client.ListMIDVAs(context.Background(), &goholdstation.ListMIDVAsRequest{})
	if err != nil {
		t.Fatalf("ListMIDVAs: %v", err)
	}

	api.assertRequest(t, wantRequest{method: http.MethodGet, path: "/partners/mid/va", query: "", body: "", signed: true})

	data := resp.Data
	if data.TotalCount != 1 || data.Page != 1 || data.PageSize != 20 || len(data.VAs) != 1 {
		t.Fatalf("data = %+v", data)
	}

	account := data.VAs[0]
	if account.VAAccount != "9704361234" || account.Status != goholdstation.MIDVAStatusExpired || account.CreatedAt.Seconds != 1793243662 || account.UpdatedAt.Seconds != 1793243662 || account.ExpiredAt != nil {
		t.Errorf("account = %+v", account)
	}
}

func TestListMIDVAs_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusUnauthorized, unauthorizedBody)

	_, err := client.ListMIDVAs(context.Background(), nil)
	assertAPIError(t, err, http.StatusUnauthorized, "UNAUTHORIZED")
}
