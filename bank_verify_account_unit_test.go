package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestVerifyBankAccount_Success(t *testing.T) {
	t.Parallel()

	client, api := newFakeAPI(t, http.StatusOK, `{"message":"success","data":{"bank":{"id":17,"bank_name":"Vietcombank","short_name":"VCB","bin_code":"970436"},"account_name":"NGUYEN VAN A","account_number":"0011001234567","account_type":1}}`)

	resp, err := client.VerifyBankAccount(context.Background(), &goholdstation.VerifyBankAccountRequest{BinCode: "970436", AccountNumber: "0011001234567"})
	if err != nil {
		t.Fatalf("VerifyBankAccount: %v", err)
	}

	api.assertRequest(t, wantRequest{
		method: http.MethodPost, path: "/api/v1/payment-methods/verify", query: "",
		body: `{"bin_code":"970436","account_number":"0011001234567"}`, signed: false,
	})

	data := resp.Data
	if data.Bank == nil || data.Bank.BinCode != "970436" || data.AccountName != "NGUYEN VAN A" || data.AccountNumber != "0011001234567" || data.AccountType != 1 {
		t.Errorf("data = %+v", data)
	}
}

func TestVerifyBankAccount_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusBadRequest, badRequestBody)

	_, err := client.VerifyBankAccount(context.Background(), &goholdstation.VerifyBankAccountRequest{BinCode: "x", AccountNumber: "y"})
	assertAPIError(t, err, http.StatusBadRequest, "BAD_REQUEST")
}
