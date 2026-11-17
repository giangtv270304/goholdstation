package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestListDepositVATransactions_Success(t *testing.T) {
	t.Parallel()

	client, api := newFakeAPI(t, http.StatusOK, `{"message":"success","data":{"transactions":[{"id":"42","amount":"995000","fee":"5000","sender_name":"NGUYEN VAN A","sender_account":"0011001234567","sender_bank_bin":"970436","transfer_content":"topup","transaction_id":"tx_1","bank_transaction_id":"btx_1","bank_trace_id":"trace_1","created_at":"2026-10-29T03:14:22Z"}],"page":2,"page_size":5,"total_count":6}}`)

	resp, err := client.ListDepositVATransactions(context.Background(), &goholdstation.ListDepositVATransactionsRequest{VAID: "dva_01HQ8X2M4K", Page: 2, PageSize: 5})
	if err != nil {
		t.Fatalf("ListDepositVATransactions: %v", err)
	}

	api.assertRequest(t, wantRequest{
		method: http.MethodGet, path: "/partners/funds/deposit-vas/dva_01HQ8X2M4K/transactions",
		query: "page=2&page_size=5", body: "", signed: true,
	})

	data := resp.Data
	if data.Page != 2 || data.PageSize != 5 || data.TotalCount != 6 || len(data.Transactions) != 1 {
		t.Fatalf("data = %+v", data)
	}

	transaction := data.Transactions[0]
	if transaction.ID != "42" || transaction.Amount != "995000" || transaction.Fee != "5000" || transaction.SenderBankBin != "970436" || transaction.BankTraceID != "trace_1" || transaction.CreatedAt != "2026-10-29T03:14:22Z" {
		t.Errorf("transaction = %+v", transaction)
	}
}

func TestListDepositVATransactions_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusNotFound, `{"error":"not found","error_code":"NOT_FOUND","message":"No deposit VA with this ID"}`)

	_, err := client.ListDepositVATransactions(context.Background(), &goholdstation.ListDepositVATransactionsRequest{VAID: "dva_missing"})
	assertAPIError(t, err, http.StatusNotFound, "NOT_FOUND")
}
