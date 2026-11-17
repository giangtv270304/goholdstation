package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestListFundDepositAddresses_Success_BareAndWrapped(t *testing.T) {
	t.Parallel()

	const payload = `{"success":true,"deposit_addresses":[{"partner_id":"partner-xyz","chain_id":56,"address":"0x4444444444444444444444444444444444444444"}],"page":1,"page_size":20,"total_count":1}`

	for name, body := range map[string]string{"bare": payload, "wrapped": `{"message":"success","data":` + payload + `}`} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client, api := newFakeAPI(t, http.StatusOK, body)

			resp, err := client.ListFundDepositAddresses(context.Background(), &goholdstation.ListFundDepositAddressesRequest{ChainID: 56})
			if err != nil {
				t.Fatalf("ListFundDepositAddresses: %v", err)
			}

			api.assertRequest(t, wantRequest{method: http.MethodGet, path: "/partners/funds/deposit-addresses", query: "chain_id=56", body: "", signed: true})

			data := resp.Data
			if !data.Success || data.Page != 1 || data.PageSize != 20 || data.TotalCount != 1 || len(data.DepositAddresses) != 1 {
				t.Fatalf("data = %+v", data)
			}

			want := goholdstation.FundDepositAddress{PartnerID: "partner-xyz", ChainID: 56, Address: "0x4444444444444444444444444444444444444444"}
			if data.DepositAddresses[0] != want {
				t.Errorf("address = %+v, want %+v", data.DepositAddresses[0], want)
			}
		})
	}
}

func TestListFundDepositAddresses_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusUnauthorized, unauthorizedBody)

	_, err := client.ListFundDepositAddresses(context.Background(), nil)
	assertAPIError(t, err, http.StatusUnauthorized, "UNAUTHORIZED")
}
