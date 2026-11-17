package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestListDepositVAs_Success_PagedObjectAndPlainArray(t *testing.T) {
	t.Parallel()

	bodies := map[string]string{
		"paged object": `{"message":"success","data":{"deposit_vas":[` + depositVAJSON + `],"page":1,"page_size":20,"total_count":1}}`,
		"plain array":  `{"message":"success","data":[` + depositVAJSON + `]}`,
	}

	for name, body := range bodies {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client, api := newFakeAPI(t, http.StatusOK, body)

			resp, err := client.ListDepositVAs(context.Background(), &goholdstation.ListDepositVAsRequest{
				Status: goholdstation.DepositVAStatusFilterActive, Page: 1, PageSize: 20,
			})
			if err != nil {
				t.Fatalf("ListDepositVAs: %v", err)
			}

			api.assertRequest(t, wantRequest{method: http.MethodGet, path: "/partners/funds/deposit-vas", query: "page=1&page_size=20&status=1", body: "", signed: true})

			if len(resp.Data.DepositVAs) != 1 || resp.Data.DepositVAs[0].ID != "dva_01HQ8X2M4K" || resp.Data.DepositVAs[0].Bank.ShortName != "VCB" {
				t.Fatalf("data = %+v", resp.Data)
			}

			if name == "paged object" && (resp.Data.Page != 1 || resp.Data.PageSize != 20 || resp.Data.TotalCount != 1) {
				t.Errorf("paging = %+v", resp.Data)
			}
		})
	}
}

func TestListDepositVAs_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusUnauthorized, unauthorizedBody)

	_, err := client.ListDepositVAs(context.Background(), nil)
	assertAPIError(t, err, http.StatusUnauthorized, "UNAUTHORIZED")
}
