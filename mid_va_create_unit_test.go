package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestCreateMIDVA_Success_BothTimestampForms(t *testing.T) {
	t.Parallel()

	bodies := map[string]string{
		"timestamp objects": `{"message":"success","data":{"id":"va_1","partner_id":"partner-xyz","va_account":"9704361234","va_name":"SDK SHOP","va_bank":"BIDV","remark":"sdk","status":1,"created_at":{"seconds":1793243662,"nanos":0},"updated_at":{"seconds":1793243662,"nanos":0},"expired_at":{"seconds":1793330062,"nanos":0}}}`,
		"rfc 3339 strings":  `{"message":"success","data":{"id":"va_1","partner_id":"partner-xyz","va_account":"9704361234","va_name":"SDK SHOP","va_bank":"BIDV","remark":"sdk","status":1,"created_at":"2026-10-29T03:14:22Z","updated_at":"2026-10-29T03:14:22Z","expired_at":"2026-10-30T03:14:22Z"}}`,
	}

	for name, body := range bodies {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client, api := newFakeAPI(t, http.StatusOK, body)
			expiration := int64(1793330062)

			resp, err := client.CreateMIDVA(context.Background(), &goholdstation.CreateMIDVARequest{VAName: "SDK SHOP", Remark: "sdk", VAExpirationTime: &expiration})
			if err != nil {
				t.Fatalf("CreateMIDVA: %v", err)
			}

			api.assertRequest(t, wantRequest{
				method: http.MethodPost, path: "/partners/mid/va", query: "",
				body: `{"va_name":"SDK SHOP","remark":"sdk","va_expiration_time":1793330062}`, signed: true,
			})

			account := resp.Data
			if account.ID != "va_1" || account.VAAccount != "9704361234" || account.VABank != "BIDV" || account.Status != goholdstation.MIDVAStatusActive {
				t.Errorf("account = %+v", account)
			}

			if account.CreatedAt == nil || account.CreatedAt.Seconds != 1793243662 || account.ExpiredAt == nil || account.ExpiredAt.Seconds != 1793330062 {
				t.Errorf("created_at = %+v, expired_at = %+v", account.CreatedAt, account.ExpiredAt)
			}
		})
	}
}

func TestCreateMIDVA_NilExpirationIsSentAsNull(t *testing.T) {
	t.Parallel()

	client, api := newFakeAPI(t, http.StatusOK, `{"message":"success","data":{"id":"va_2"}}`)

	if _, err := client.CreateMIDVA(context.Background(), &goholdstation.CreateMIDVARequest{VAName: "SDK SHOP", Remark: "sdk"}); err != nil {
		t.Fatalf("CreateMIDVA: %v", err)
	}

	api.assertRequest(t, wantRequest{
		method: http.MethodPost, path: "/partners/mid/va", query: "",
		body: `{"va_name":"SDK SHOP","remark":"sdk","va_expiration_time":null}`, signed: true,
	})
}

func TestCreateMIDVA_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusBadRequest, badRequestBody)

	_, err := client.CreateMIDVA(context.Background(), &goholdstation.CreateMIDVARequest{})
	assertAPIError(t, err, http.StatusBadRequest, "BAD_REQUEST")
}
