package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestSetWebhookURL_Success_BareAndWrapped(t *testing.T) {
	t.Parallel()

	const payload = `{"webhook_url":"https://yourapp.com/api/webhooks/fund-balance"}`

	for name, body := range map[string]string{"bare": payload, "wrapped": `{"message":"success","data":` + payload + `}`} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client, api := newFakeAPI(t, http.StatusOK, body)

			resp, err := client.SetWebhookURL(context.Background(), &goholdstation.SetWebhookURLRequest{WebhookURL: "https://yourapp.com/api/webhooks/fund-balance"})
			if err != nil {
				t.Fatalf("SetWebhookURL: %v", err)
			}

			api.assertRequest(t, wantRequest{
				method: http.MethodPost, path: "/partners/webhook-url", query: "",
				body: `{"webhook_url":"https://yourapp.com/api/webhooks/fund-balance"}`, signed: true,
			})

			if resp.Data.WebhookURL != "https://yourapp.com/api/webhooks/fund-balance" {
				t.Errorf("webhook_url = %q", resp.Data.WebhookURL)
			}
		})
	}
}

func TestSetWebhookURL_EmptyStringIsSentToClear(t *testing.T) {
	t.Parallel()

	client, api := newFakeAPI(t, http.StatusOK, `{"webhook_url":""}`)

	if _, err := client.SetWebhookURL(context.Background(), &goholdstation.SetWebhookURLRequest{WebhookURL: ""}); err != nil {
		t.Fatalf("SetWebhookURL: %v", err)
	}

	api.assertRequest(t, wantRequest{method: http.MethodPost, path: "/partners/webhook-url", query: "", body: `{"webhook_url":""}`, signed: true})
}

func TestSetWebhookURL_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusBadRequest, badRequestBody)

	_, err := client.SetWebhookURL(context.Background(), &goholdstation.SetWebhookURLRequest{WebhookURL: "ftp://x"})
	assertAPIError(t, err, http.StatusBadRequest, "BAD_REQUEST")
}
