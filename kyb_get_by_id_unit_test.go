package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestGetKYBByID_Success(t *testing.T) {
	t.Parallel()

	client, api := newFakeAPI(t, http.StatusOK, `{"message":"success","data":{"kyb":`+kybJSON+`}}`)

	resp, err := client.GetKYBByID(context.Background(), &goholdstation.GetKYBByIDRequest{ID: "kyb_01HQ"})
	if err != nil {
		t.Fatalf("GetKYBByID: %v", err)
	}

	api.assertRequest(t, wantRequest{method: http.MethodGet, path: "/partners/kyb/kyb_01HQ", query: "", body: "", signed: true})

	assertKYB(t, resp.Data.KYB)
}

func TestGetKYBByID_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusNotFound, `{"error":"NOT_FOUND","message":"KYB record not found"}`)

	_, err := client.GetKYBByID(context.Background(), &goholdstation.GetKYBByIDRequest{ID: "kyb_missing"})
	assertAPIError(t, err, http.StatusNotFound, "NOT_FOUND")
}
