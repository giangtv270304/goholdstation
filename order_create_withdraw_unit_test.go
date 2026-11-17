package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestCreateWithdrawOrder_Success(t *testing.T) {
	t.Parallel()

	client, api := newFakeAPI(t, http.StatusOK, `{"message":"success","data":{"success":true,"code":0,"error":"","order":{"id":"ord_p1","order_type":2,"state":1,"processing_state":0,"amount":"2575000","currency":"vnd","outcome":"100.000001","outcome_currency":"token","rate":"25750","total_fee_vnd":11000,"purpose":"salary","payment_info":{"bank_id":17,"full_name":"NGUYEN VAN A","account_type":1,"account_number":"0011001234567"},"expired_at":{"seconds":1700000900,"nanos":0}}}}`)

	resp, err := client.CreateWithdrawOrder(context.Background(), &goholdstation.CreateWithdrawOrderRequest{
		Amount: "2575000", Currency: "vnd", TokenSymbol: "USDT",
		PaymentInfo: &goholdstation.PaymentInfo{BankID: 17, FullName: "NGUYEN VAN A", AccountType: 1, AccountNumber: "0011001234567"},
		Callback:    "https://yourapp.com/cb", ReferenceNo: "ref-p1", IdempotencyKey: "idem-1", KYCID: "kyc-1", Purpose: "salary",
	})
	if err != nil {
		t.Fatalf("CreateWithdrawOrder: %v", err)
	}

	api.assertRequest(t, wantRequest{
		method: http.MethodPost, path: "/partners/orders/withdraw", query: "",
		body:   `{"amount":"2575000","currency":"vnd","token_symbol":"USDT","payment_info":{"bank_id":17,"full_name":"NGUYEN VAN A","account_type":1,"account_number":"0011001234567"},"callback":"https://yourapp.com/cb","reference_no":"ref-p1","idempotency_key":"idem-1","kyc_id":"kyc-1","purpose":"salary"}`,
		signed: true,
	})

	order := resp.Data.Order
	if !resp.Data.Success || order == nil || order.ID != "ord_p1" || order.State != goholdstation.OrderStateCreated {
		t.Fatalf("data = %+v order = %+v", resp.Data, order)
	}

	if order.Amount != "2575000" || order.Outcome != "100.000001" || order.TotalFeeVND != "11000" || order.Purpose != "salary" {
		t.Errorf("order = %+v", order)
	}
}

func TestCreateWithdrawOrder_BusinessFailureIsAnAPIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusOK, `{"message":"success","data":{"success":false,"code":4,"error":"duplicate request","order":null}}`)

	_, err := client.CreateWithdrawOrder(context.Background(), &goholdstation.CreateWithdrawOrderRequest{Amount: "1", IdempotencyKey: "idem-1"})
	assertAPIError(t, err, http.StatusOK, "4")
}

func TestCreateWithdrawOrder_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusBadRequest, `{"error":"bad request","error_code":"INSUFFICIENT_BALANCE","message":"fund balance is too low"}`)

	_, err := client.CreateWithdrawOrder(context.Background(), &goholdstation.CreateWithdrawOrderRequest{Amount: "1"})
	assertAPIError(t, err, http.StatusBadRequest, "INSUFFICIENT_BALANCE")
}
