package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestListKYCs_Success_PagedObjectAndPlainArray(t *testing.T) {
	t.Parallel()

	bodies := map[string]string{
		"paged object": `{"message":"success","data":{"success":true,"kycs":[` + kycJSON + `],"page":2,"page_size":10,"total_count":11}}`,
		"plain array":  `{"message":"success","data":[` + kycJSON + `]}`,
	}

	for name, body := range bodies {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client, api := newFakeAPI(t, http.StatusOK, body)

			resp, err := client.ListKYCs(context.Background(), &goholdstation.ListKYCsRequest{Page: 2, PageSize: 10})
			if err != nil {
				t.Fatalf("ListKYCs: %v", err)
			}

			api.assertRequest(t, wantRequest{method: http.MethodGet, path: "/partners/kycs", query: "page=2&page_size=10", body: "", signed: true})

			if len(resp.Data.KYCs) != 1 {
				t.Fatalf("data = %+v", resp.Data)
			}

			assertKYC(t, &resp.Data.KYCs[0])

			if name == "paged object" && (resp.Data.Page != 2 || resp.Data.PageSize != 10 || resp.Data.TotalCount != 11) {
				t.Errorf("paging = %+v", resp.Data)
			}
		})
	}
}

func TestListKYCs_EmptyPageWithoutTotalCount(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusOK, `{"message":"success","data":{"success":true,"kycs":null,"page":1,"page_size":20}}`)

	resp, err := client.ListKYCs(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListKYCs: %v", err)
	}

	if resp.Data.KYCs != nil || resp.Data.TotalCount != 0 || resp.Data.PageSize != 20 {
		t.Errorf("data = %+v", resp.Data)
	}
}

func TestListKYCs_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusUnauthorized, unauthorizedBody)

	_, err := client.ListKYCs(context.Background(), nil)
	assertAPIError(t, err, http.StatusUnauthorized, "UNAUTHORIZED")
}
