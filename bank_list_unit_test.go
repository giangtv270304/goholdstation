package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestListBanks_Success(t *testing.T) {
	t.Parallel()

	client, api := newFakeAPI(t, http.StatusOK, `{"message":"success","data":{"banks":[{"id":17,"bank_name":"Vietcombank","short_name":"VCB","bin_code":"970436","image":"https://img.example/vcb.png","created_at":{"seconds":1700000000,"nanos":0},"updated_at":{"seconds":1700000001,"nanos":0}}],"page":2,"page_size":10,"total_count":41}}`)

	resp, err := client.ListBanks(context.Background(), &goholdstation.ListBanksRequest{Page: 2, Size: 10, IDs: []int64{17, 18}})
	if err != nil {
		t.Fatalf("ListBanks: %v", err)
	}

	api.assertRequest(t, wantRequest{method: http.MethodGet, path: "/api/v1/banks", query: "ids=17%2C18&page=2&size=10", body: "", signed: false})

	data := resp.Data
	if resp.Message != "success" || len(data.Banks) != 1 || data.Page != 2 || data.PageSize != 10 || data.TotalCount != 41 {
		t.Fatalf("resp = %+v", resp)
	}

	bank := data.Banks[0]
	if bank.ID != 17 || bank.BankName != "Vietcombank" || bank.ShortName != "VCB" || bank.BinCode != "970436" || bank.CreatedAt.Seconds != 1700000000 {
		t.Errorf("bank = %+v", bank)
	}
}

func TestListBanks_NilRequestSendsNoQuery(t *testing.T) {
	t.Parallel()

	client, api := newFakeAPI(t, http.StatusOK, `{"message":"success","data":{"banks":null,"page":1,"page_size":20,"total_count":0}}`)

	resp, err := client.ListBanks(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListBanks: %v", err)
	}

	api.assertRequest(t, wantRequest{method: http.MethodGet, path: "/api/v1/banks", query: "", body: "", signed: false})

	if resp.Data.Banks != nil {
		t.Errorf("banks = %+v, want nil", resp.Data.Banks)
	}
}

func TestListBanks_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusInternalServerError, `{"error":"internal","error_code":"INTERNAL","message":"boom"}`)

	_, err := client.ListBanks(context.Background(), nil)
	assertAPIError(t, err, http.StatusInternalServerError, "INTERNAL")
}
