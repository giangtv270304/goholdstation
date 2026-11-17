package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestDecodeQR_Success_BareAndWrapped(t *testing.T) {
	t.Parallel()

	const payload = `{"data":{"bank_bin_code":"970436","bank_account_number":"0011001234567","amount":150000,"purpose":"coffee"},"error":"","error_code":0,"message":"ok"}`

	bodies := map[string]string{
		"bare":                  payload,
		"wrapped":               `{"message":"success","data":` + payload + `}`,
		"bare, empties omitted": `{"data":{"bank_bin_code":"970436","bank_account_number":"0011001234567","amount":150000,"purpose":"coffee"}}`,
		"wrapped, only data":    `{"message":"success","data":{"data":{"bank_bin_code":"970436","bank_account_number":"0011001234567","amount":150000,"purpose":"coffee"}}}`,
	}

	for name, body := range bodies {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client, api := newFakeAPI(t, http.StatusOK, body)

			resp, err := client.DecodeQR(context.Background(), &goholdstation.DecodeQRRequest{QRString: "000201010212"})
			if err != nil {
				t.Fatalf("DecodeQR: %v", err)
			}

			api.assertRequest(t, wantRequest{
				method: http.MethodPost, path: "/partners/payment-methods/qr/decode", query: "",
				body: `{"qr_string":"000201010212"}`, signed: true,
			})

			qr := resp.Data.Data
			if qr == nil || qr.BankBinCode != "970436" || qr.BankAccountNumber != "0011001234567" || qr.Amount != "150000" || qr.Purpose != "coffee" {
				t.Errorf("qr = %+v", qr)
			}

			if resp.Data.ErrorCode != 0 {
				t.Errorf("error code = %d, want 0", resp.Data.ErrorCode)
			}
		})
	}
}

func TestDecodeQR_FailedDecodeIsReturnedNotAnError(t *testing.T) {
	t.Parallel()

	const payload = `{"data":null,"error":"image download failed","error_code":2,"message":"failed"}`

	for name, body := range map[string]string{"bare": payload, "wrapped": `{"message":"success","data":` + payload + `}`} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client, api := newFakeAPI(t, http.StatusOK, body)

			resp, err := client.DecodeQR(context.Background(), &goholdstation.DecodeQRRequest{QRImageURL: "https://img.example/qr.png"})
			if err != nil {
				t.Fatalf("DecodeQR: %v", err)
			}

			api.assertRequest(t, wantRequest{
				method: http.MethodPost, path: "/partners/payment-methods/qr/decode", query: "",
				body: `{"qr_image_url":"https://img.example/qr.png"}`, signed: true,
			})

			if resp.Data.Data != nil || resp.Data.ErrorCode != 2 || resp.Data.Error != "image download failed" || resp.Data.Message != "failed" {
				t.Errorf("data = %+v", resp.Data)
			}
		})
	}
}

func TestDecodeQR_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusUnauthorized, unauthorizedBody)

	_, err := client.DecodeQR(context.Background(), &goholdstation.DecodeQRRequest{QRString: "x"})
	assertAPIError(t, err, http.StatusUnauthorized, "UNAUTHORIZED")
}
