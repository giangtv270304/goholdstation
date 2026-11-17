package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestGetWebhookURL_Success_BareAndWrapped(t *testing.T) {
	t.Parallel()

	const payload = `{"webhook_url":"https://yourapp.com/api/webhooks/fund-balance"}`

	for name, body := range map[string]string{"bare": payload, "wrapped": `{"message":"success","data":` + payload + `}`} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client, api := newFakeAPI(t, http.StatusOK, body)

			resp, err := client.GetWebhookURL(context.Background(), &goholdstation.GetWebhookURLRequest{})
			if err != nil {
				t.Fatalf("GetWebhookURL: %v", err)
			}

			api.assertRequest(t, wantRequest{method: http.MethodGet, path: "/partners/webhook-url", query: "", body: "", signed: true})

			if resp.Data.WebhookURL != "https://yourapp.com/api/webhooks/fund-balance" {
				t.Errorf("webhook_url = %q", resp.Data.WebhookURL)
			}
		})
	}
}

func TestGetWebhookURL_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusUnauthorized, unauthorizedBody)

	_, err := client.GetWebhookURL(context.Background(), nil)
	assertAPIError(t, err, http.StatusUnauthorized, "UNAUTHORIZED")
}
