package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestRevokeMIDVA_Success_CamelAndSnakeCase(t *testing.T) {
	t.Parallel()

	bodies := map[string]string{
		"camelCase":  `{"message":"success","data":{"merchantId":"M001","vaAccount":"9704361234","vaStatus":"4"}}`,
		"snake_case": `{"message":"success","data":{"merchant_id":"M001","va_account":"9704361234","va_status":"4"}}`,
	}

	for name, body := range bodies {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client, api := newFakeAPI(t, http.StatusOK, body)

			resp, err := client.RevokeMIDVA(context.Background(), &goholdstation.RevokeMIDVARequest{VAAccount: "9704361234"})
			if err != nil {
				t.Fatalf("RevokeMIDVA: %v", err)
			}

			api.assertRequest(t, wantRequest{method: http.MethodDelete, path: "/partners/mid/va/9704361234", query: "", body: "", signed: true})

			want := goholdstation.RevokeMIDVAResponse{MerchantID: "M001", VAAccount: "9704361234", VAStatus: "4"}
			if resp.Data != want {
				t.Errorf("data = %+v, want %+v", resp.Data, want)
			}
		})
	}
}

func TestRevokeMIDVA_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusNotFound, notFoundBody)

	_, err := client.RevokeMIDVA(context.Background(), &goholdstation.RevokeMIDVARequest{VAAccount: "0"})
	assertAPIError(t, err, http.StatusNotFound, "NOT_FOUND")
}
