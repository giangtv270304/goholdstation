package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestUpdateKYC_Success_BothDocumentedForms(t *testing.T) {
	t.Parallel()

	for name, body := range kycBodies {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client, api := newFakeAPI(t, http.StatusOK, body)

			resp, err := client.UpdateKYC(context.Background(), &goholdstation.UpdateKYCRequest{
				ID:  "cust-1",
				KYC: goholdstation.KYCInput{Name: "NGUYEN VAN A", Phone: "+84900000000", IDType: goholdstation.KYCIDTypeCitizenID},
			})
			if err != nil {
				t.Fatalf("UpdateKYC: %v", err)
			}

			api.assertRequest(t, wantRequest{
				method: http.MethodPut, path: "/partners/kycs/cust-1", query: "",
				body: `{"name":"NGUYEN VAN A","phone":"+84900000000","id_type":1}`, signed: true,
			})

			assertKYC(t, resp.Data.KYC)
		})
	}
}

func TestUpdateKYC_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusNotFound, `{"error":"not found","error_code":"KYC_NOT_FOUND","message":"KYC not found"}`)

	_, err := client.UpdateKYC(context.Background(), &goholdstation.UpdateKYCRequest{ID: "cust-missing"})
	assertAPIError(t, err, http.StatusNotFound, "KYC_NOT_FOUND")
}
