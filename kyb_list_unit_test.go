package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestListKYBs_Success(t *testing.T) {
	t.Parallel()

	client, api := newFakeAPI(t, http.StatusOK, `{"message":"success","data":{"success":true,"kybs":[`+kybJSON+`],"page":1,"page_size":20,"total_count":1}}`)

	resp, err := client.ListKYBs(context.Background(), &goholdstation.ListKYBsRequest{
		ReferenceID: "merchant 7", SubjectType: goholdstation.KYBSubjectTypeSubMerchant, Status: goholdstation.KYBStatusPending, Page: 1, PageSize: 20,
	})
	if err != nil {
		t.Fatalf("ListKYBs: %v", err)
	}

	api.assertRequest(t, wantRequest{
		method: http.MethodGet, path: "/partners/kyb",
		query: "page=1&page_size=20&reference_id=merchant+7&status=1&subject_type=2", body: "", signed: true,
	})

	data := resp.Data
	if data.Page != 1 || data.PageSize != 20 || data.TotalCount != 1 || len(data.KYBs) != 1 {
		t.Fatalf("data = %+v", data)
	}

	assertKYB(t, &data.KYBs[0])
}

func TestListKYBs_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusUnauthorized, unauthorizedBody)

	_, err := client.ListKYBs(context.Background(), nil)
	assertAPIError(t, err, http.StatusUnauthorized, "UNAUTHORIZED")
}
