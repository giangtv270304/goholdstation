package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestCreateWithdrawalOrder_Success(t *testing.T) {
	t.Parallel()

	client, api := newFakeAPI(t, http.StatusOK, `{"message":"success","data":{"order":{"id":"ord_w1","order_type":2,"state":2,"processing_state":20,"amount":"100.123456789012345678","currency":"token","outcome":"2575000","outcome_currency":"vnd","rate":"25750","recipient":"0x2222222222222222222222222222222222222222","refund_address":"0x3333333333333333333333333333333333333333","payment_info":{"bank_id":17,"full_name":"NGUYEN VAN A","account_type":1,"account_number":"0011001234567","transfer_note":"thanks"}}}}`)

	resp, err := client.CreateWithdrawalOrder(context.Background(), &goholdstation.CreateWithdrawalOrderRequest{
		Amount: "100.123456789012345678", Currency: "token", ChainID: 56, TokenAddress: "0xabc",
		PaymentInfo: &goholdstation.PaymentInfo{BankID: 17, FullName: "NGUYEN VAN A", AccountType: 1, AccountNumber: "0011001234567", TransferNote: "thanks"},
		Callback:    "https://yourapp.com/cb", ReferenceNo: "ref-w1", Buffer: decimalPtr("0.01"), KYCID: "kyc-1",
		RefundAddress: "0x3333333333333333333333333333333333333333",
	})
	if err != nil {
		t.Fatalf("CreateWithdrawalOrder: %v", err)
	}

	api.assertRequest(t, wantRequest{
		method: http.MethodPost, path: "/partners/orders/withdrawal", query: "",
		body:   `{"amount":"100.123456789012345678","currency":"token","chain_id":56,"token_address":"0xabc","payment_info":{"bank_id":17,"full_name":"NGUYEN VAN A","account_type":1,"account_number":"0011001234567","transfer_note":"thanks"},"callback":"https://yourapp.com/cb","reference_no":"ref-w1","buffer":0.01,"kyc_id":"kyc-1","refund_address":"0x3333333333333333333333333333333333333333"}`,
		signed: false,
	})

	order := resp.Data.Order
	if order == nil || order.OrderType != goholdstation.OrderTypeOfframp || order.ProcessingState != goholdstation.OrderProcessingStateOfframpCryptoPending {
		t.Fatalf("order = %+v", order)
	}

	if order.Amount != "100.123456789012345678" || order.Outcome != "2575000" || order.Recipient != "0x2222222222222222222222222222222222222222" {
		t.Errorf("order = %+v", order)
	}

	if order.PaymentInfo == nil || order.PaymentInfo.BankID != 17 || order.PaymentInfo.TransferNote != "thanks" {
		t.Errorf("payment info = %+v", order.PaymentInfo)
	}
}

func TestCreateWithdrawalOrder_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusBadRequest, `{"error":"bad request","error_code":"400105","message":"order type not available"}`)

	_, err := client.CreateWithdrawalOrder(context.Background(), &goholdstation.CreateWithdrawalOrderRequest{Amount: "1"})
	assertAPIError(t, err, http.StatusBadRequest, "400105")
}
