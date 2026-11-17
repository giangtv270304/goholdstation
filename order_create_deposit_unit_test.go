package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

const depositOrderJSON = `{"id":"3c1f6f8e-6d0a-4f0e-9d7a-0b6d2f1a9c11","code":100001,"order_type":1,"state":2,"processing_state":10,"amount":"1000000","currency":"vnd","outcome":"38.61","outcome_currency":"token","rate":"25900.5","original_rate":"25850","fees":{"service_fee":5000,"min_fee":0.01},"total_fee_vnd":5000.25,"token_address":"0x55d398326f99059ff775485246999027b3197955","chain_id":56,"recipient":"0x1111111111111111111111111111111111111111","reference_no":"ref-1","partner_id":"partner-xyz","callback":"https://yourapp.com/cb","pay_data":{"bank_account_name":"HOLDSTATION","bank_account_number":"9001","bank_bin_code":"970436","bank_name":"Vietcombank","bank_short_code":"VCB","payment_link":"https://pay.example/l","qr_code":"000201","qr_image":"https://img.example/qr.png","external_id":"ext_1"},"payment_info":null,"error":null,"created_at":{"seconds":1700000000,"nanos":0},"updated_at":{"seconds":1700000000,"nanos":0},"expired_at":{"seconds":1700000900,"nanos":0}}`

func TestCreateDepositOrder_Success(t *testing.T) {
	t.Parallel()

	client, api := newFakeAPI(t, http.StatusOK, `{"message":"success","data":{"order":`+depositOrderJSON+`}}`)

	resp, err := client.CreateDepositOrder(context.Background(), &goholdstation.CreateDepositOrderRequest{
		Amount: "1000000", Currency: "vnd", ChainID: 56,
		TokenAddress: "0x55d398326f99059ff775485246999027b3197955",
		Recipient:    "0x1111111111111111111111111111111111111111",
		Callback:     "https://yourapp.com/cb", ReferenceNo: "ref-1", Buffer: decimalPtr("0.005"), KYCID: "kyc-1",
	})
	if err != nil {
		t.Fatalf("CreateDepositOrder: %v", err)
	}

	api.assertRequest(t, wantRequest{
		method: http.MethodPost, path: "/partners/orders/deposit", query: "",
		body:   `{"amount":"1000000","currency":"vnd","chain_id":56,"token_address":"0x55d398326f99059ff775485246999027b3197955","recipient":"0x1111111111111111111111111111111111111111","callback":"https://yourapp.com/cb","reference_no":"ref-1","buffer":0.005,"kyc_id":"kyc-1"}`,
		signed: false,
	})

	order := resp.Data.Order
	if order == nil || order.ID != "3c1f6f8e-6d0a-4f0e-9d7a-0b6d2f1a9c11" || order.Code != 100001 {
		t.Fatalf("order = %+v", order)
	}

	if order.OrderType != goholdstation.OrderTypeOnramp || order.State != goholdstation.OrderStateProcessing || order.ProcessingState != goholdstation.OrderProcessingStateOnrampFiatPending {
		t.Errorf("order states = %d %d %d", order.OrderType, order.State, order.ProcessingState)
	}

	if order.Amount != "1000000" || order.Outcome != "38.61" || order.Rate != "25900.5" || order.OriginalRate != "25850" || order.TotalFeeVND != "5000.25" {
		t.Errorf("order money fields = %+v", order)
	}

	if order.PayData == nil || order.PayData.QRCode != "000201" || order.PayData.BankAccountNumber != "9001" || order.PaymentInfo != nil || order.Error != nil {
		t.Errorf("pay data = %+v, payment info = %+v, error = %+v", order.PayData, order.PaymentInfo, order.Error)
	}

	if order.ExpiredAt == nil || order.ExpiredAt.Seconds != 1700000900 {
		t.Errorf("expired_at = %+v", order.ExpiredAt)
	}
}

func TestCreateDepositOrder_OptionalFieldsAreLeftOut(t *testing.T) {
	t.Parallel()

	client, api := newFakeAPI(t, http.StatusOK, `{"message":"success","data":{"order":{"id":"ord_2"}}}`)

	_, err := client.CreateDepositOrder(context.Background(), &goholdstation.CreateDepositOrderRequest{
		Amount: "50000.5", Currency: "vnd", ChainID: 56, TokenAddress: "0xabc", Recipient: "0xdef", Callback: "https://yourapp.com/cb",
	})
	if err != nil {
		t.Fatalf("CreateDepositOrder: %v", err)
	}

	api.assertRequest(t, wantRequest{
		method: http.MethodPost, path: "/partners/orders/deposit", query: "",
		body: `{"amount":"50000.5","currency":"vnd","chain_id":56,"token_address":"0xabc","recipient":"0xdef","callback":"https://yourapp.com/cb"}`, signed: false,
	})
}

func TestCreateDepositOrder_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusBadRequest, `{"error":"bad request","error_code":"400110","message":"under maintenance"}`)

	_, err := client.CreateDepositOrder(context.Background(), &goholdstation.CreateDepositOrderRequest{Amount: "1"})
	assertAPIError(t, err, http.StatusBadRequest, "400110")
}
