package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestCreateWholesaleApplication_Success(t *testing.T) {
	t.Parallel()

	client, api := newFakeAPI(t, http.StatusOK, `{"message":"success","data":{"application":{"id":"app_1","user_id":"usr_1","email":"ops@example.com","purpose":"otc","default_rate":0.01,"status":1,"submitted_at":{"seconds":1700000000,"nanos":0}}}}`)

	resp, err := client.CreateWholesaleApplication(context.Background(), &goholdstation.CreateWholesaleApplicationRequest{DefaultRate: decimalPtr("0.01")})
	if err != nil {
		t.Fatalf("CreateWholesaleApplication: %v", err)
	}

	api.assertRequest(t, wantRequest{method: http.MethodPost, path: "/api/v1/wholesale/applications", query: "", body: `{"default_rate":0.01}`, signed: false})

	application := resp.Data.Application
	if application == nil || application.ID != "app_1" || application.DefaultRate != "0.01" || application.Status != 1 || application.SubmittedAt.Seconds != 1700000000 || application.ReviewedAt != nil {
		t.Errorf("application = %+v", application)
	}
}

func TestCreateWholesaleApplication_NilRateIsSentAsNull(t *testing.T) {
	t.Parallel()

	client, api := newFakeAPI(t, http.StatusOK, `{"message":"success","data":{"application":{"id":"app_2"}}}`)

	if _, err := client.CreateWholesaleApplication(context.Background(), &goholdstation.CreateWholesaleApplicationRequest{}); err != nil {
		t.Fatalf("CreateWholesaleApplication: %v", err)
	}

	api.assertRequest(t, wantRequest{method: http.MethodPost, path: "/api/v1/wholesale/applications", query: "", body: `{"default_rate":null}`, signed: false})
}

func TestCreateWholesaleApplication_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusUnauthorized, unauthorizedBody)

	_, err := client.CreateWholesaleApplication(context.Background(), &goholdstation.CreateWholesaleApplicationRequest{})
	assertAPIError(t, err, http.StatusUnauthorized, "UNAUTHORIZED")
}
