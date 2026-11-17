package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestGetKYCByID_Success_BothDocumentedForms(t *testing.T) {
	t.Parallel()

	for name, body := range kycBodies {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client, api := newFakeAPI(t, http.StatusOK, body)

			resp, err := client.GetKYCByID(context.Background(), &goholdstation.GetKYCByIDRequest{ID: "cust-1"})
			if err != nil {
				t.Fatalf("GetKYCByID: %v", err)
			}

			api.assertRequest(t, wantRequest{method: http.MethodGet, path: "/partners/kycs/cust-1", query: "", body: "", signed: true})

			assertKYC(t, resp.Data.KYC)
		})
	}
}

func TestGetKYCByID_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusNotFound, `{"error":"not found","error_code":"KYC_NOT_FOUND","message":"KYC not found"}`)

	_, err := client.GetKYCByID(context.Background(), &goholdstation.GetKYCByIDRequest{ID: "cust-missing"})
	if !goholdstation.IsNotFound(err) {
		t.Errorf("error = %v, want IsNotFound", err)
	}

	assertAPIError(t, err, http.StatusNotFound, "KYC_NOT_FOUND")
}
