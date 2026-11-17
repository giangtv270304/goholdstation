package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestConvertFund_Success_BareAndWrapped(t *testing.T) {
	t.Parallel()

	const payload = `{"success":true,"code":0,"error":"","rate":"1","converted_amount":"100.000000000000000001","received_amount":"99.900000000000000001","fee":"0.1","from_fund":{"id":1,"token":"USDT","balance":"900.5"},"to_fund":{"id":2,"token":"USDC","balance":"99.900000000000000001"},"from_log":{"id":41,"partner_fund_id":1,"change_type":4,"change_amount":"-100.000000000000000001","fee":"0","pre_balance":"1000.500000000000000001","post_balance":"900.5","metadata":{"source":"convert","reference_id":"cv_1","convert_ref":"cv_1","from_token":"USDT","to_token":"USDC","rate":"1"},"caused_at":{"seconds":1700000000,"nanos":0}},"to_log":{"id":42,"partner_fund_id":2,"change_type":5,"change_amount":"99.900000000000000001","fee":"0.1","pre_balance":"0","post_balance":"99.900000000000000001"}}`

	for name, body := range map[string]string{"bare": payload, "wrapped": `{"message":"success","data":` + payload + `}`} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client, api := newFakeAPI(t, http.StatusOK, body)

			resp, err := client.ConvertFund(context.Background(), &goholdstation.ConvertFundRequest{
				FromToken: "USDT", ToToken: "USDC", Amount: "100.000000000000000001", IdempotencyKey: "idem-cv-1",
			})
			if err != nil {
				t.Fatalf("ConvertFund: %v", err)
			}

			api.assertRequest(t, wantRequest{
				method: http.MethodPost, path: "/partners/funds/convert", query: "",
				body: `{"from_token":"USDT","to_token":"USDC","amount":"100.000000000000000001","idempotency_key":"idem-cv-1"}`, signed: true,
			})

			data := resp.Data
			if data.Rate != "1" || data.ConvertedAmount != "100.000000000000000001" || data.ReceivedAmount != "99.900000000000000001" || data.Fee != "0.1" {
				t.Errorf("data = %+v", data)
			}

			if data.FromFund == nil || data.FromFund.Balance != "900.5" || data.ToFund == nil || data.ToFund.Token != "USDC" {
				t.Errorf("funds = %+v %+v", data.FromFund, data.ToFund)
			}

			if data.FromLog == nil || data.FromLog.ChangeType != goholdstation.FundChangeTypeConvertOut || data.FromLog.ChangeAmount != "-100.000000000000000001" ||
				data.ToLog == nil || data.ToLog.ChangeType != goholdstation.FundChangeTypeConvertIn {
				t.Errorf("logs = %+v %+v", data.FromLog, data.ToLog)
			}
		})
	}
}

func TestConvertFund_WholeBalanceLeavesAmountOut(t *testing.T) {
	t.Parallel()

	client, api := newFakeAPI(t, http.StatusOK, `{"success":true,"code":0,"error":"","rate":"1"}`)

	if _, err := client.ConvertFund(context.Background(), &goholdstation.ConvertFundRequest{FromToken: "USDT", ToToken: "USD1"}); err != nil {
		t.Fatalf("ConvertFund: %v", err)
	}

	api.assertRequest(t, wantRequest{method: http.MethodPost, path: "/partners/funds/convert", query: "", body: `{"from_token":"USDT","to_token":"USD1"}`, signed: true})
}

func TestConvertFund_BusinessFailureIsAnAPIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusOK, `{"success":false,"code":3,"error":"insufficient balance"}`)

	_, err := client.ConvertFund(context.Background(), &goholdstation.ConvertFundRequest{FromToken: "USDT", ToToken: "USDC"})
	assertAPIError(t, err, http.StatusOK, "3")
}

func TestConvertFund_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusBadRequest, badRequestBody)

	_, err := client.ConvertFund(context.Background(), &goholdstation.ConvertFundRequest{FromToken: "USDT", ToToken: "USDT"})
	assertAPIError(t, err, http.StatusBadRequest, "BAD_REQUEST")
}
