package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

const depositVAJSON = `{"id":"dva_01HQ8X2M4K","account_number":"900112345678","account_name":"SDK SHOP","bank_id":17,"bank":{"id":17,"bin_code":"970436","short_name":"VCB","bank_name":"Vietcombank","image":"https://img.example/vcb.png"},"status":"ACTIVE","created_at":"2026-10-29T03:14:22Z"}`

func TestCreateDepositVA_Success(t *testing.T) {
	t.Parallel()

	client, api := newFakeAPI(t, http.StatusOK, `{"message":"success","data":`+depositVAJSON+`}`)

	resp, err := client.CreateDepositVA(context.Background(), &goholdstation.CreateDepositVARequest{VAName: "SDK SHOP"})
	if err != nil {
		t.Fatalf("CreateDepositVA: %v", err)
	}

	api.assertRequest(t, wantRequest{method: http.MethodPost, path: "/partners/funds/deposit-vas", query: "", body: `{"va_name":"SDK SHOP"}`, signed: true})

	account := resp.Data
	if account.ID != "dva_01HQ8X2M4K" || account.AccountNumber != "900112345678" || account.AccountName != "SDK SHOP" || account.BankID != 17 {
		t.Errorf("account = %+v", account)
	}

	if account.Bank == nil || account.Bank.BinCode != "970436" || account.Status != goholdstation.DepositVAStatusActive || account.CreatedAt != "2026-10-29T03:14:22Z" {
		t.Errorf("account = %+v bank = %+v", account, account.Bank)
	}
}

func TestCreateDepositVA_WithoutNameSendsNoBody(t *testing.T) {
	t.Parallel()

	for name, req := range map[string]*goholdstation.CreateDepositVARequest{"nil request": nil, "empty name": {}} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client, api := newFakeAPI(t, http.StatusOK, `{"message":"success","data":`+depositVAJSON+`}`)

			if _, err := client.CreateDepositVA(context.Background(), req); err != nil {
				t.Fatalf("CreateDepositVA: %v", err)
			}

			api.assertRequest(t, wantRequest{method: http.MethodPost, path: "/partners/funds/deposit-vas", query: "", body: "", signed: true})
		})
	}
}

func TestCreateDepositVA_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusBadRequest, badRequestBody)

	_, err := client.CreateDepositVA(context.Background(), &goholdstation.CreateDepositVARequest{VAName: ""})
	assertAPIError(t, err, http.StatusBadRequest, "BAD_REQUEST")
}
