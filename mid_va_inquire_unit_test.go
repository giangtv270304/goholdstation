package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestInquireMIDVA_Success_CamelAndSnakeCase(t *testing.T) {
	t.Parallel()

	bodies := map[string]string{
		"camelCase":  `{"message":"success","data":{"merchantId":"M001","vaAccount":"9704361234","vaName":"SDK SHOP","vaBank":"BIDV","transactions":[{"transId":9007199254740993,"amount":150000,"timePaid":1700000000,"cashinId":77,"remark":"coffee"}]}}`,
		"snake_case": `{"message":"success","data":{"merchant_id":"M001","va_account":"9704361234","va_name":"SDK SHOP","va_bank":"BIDV","transactions":[{"trans_id":9007199254740993,"amount":150000,"time_paid":1700000000,"cashin_id":77,"remark":"coffee"}]}}`,
	}

	for name, body := range bodies {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client, api := newFakeAPI(t, http.StatusOK, body)

			resp, err := client.InquireMIDVA(context.Background(), &goholdstation.InquireMIDVARequest{VAAccount: "9704361234"})
			if err != nil {
				t.Fatalf("InquireMIDVA: %v", err)
			}

			api.assertRequest(t, wantRequest{method: http.MethodGet, path: "/partners/mid/va/9704361234/inquiry", query: "", body: "", signed: true})

			data := resp.Data
			if data.MerchantID != "M001" || data.VAAccount != "9704361234" || data.VAName != "SDK SHOP" || data.VABank != "BIDV" || len(data.Transactions) != 1 {
				t.Fatalf("data = %+v", data)
			}

			want := goholdstation.MIDVATransaction{TransID: 9007199254740993, Amount: "150000", TimePaid: 1700000000, CashinID: 77, Remark: "coffee"}
			if data.Transactions[0] != want {
				t.Errorf("transaction = %+v, want %+v", data.Transactions[0], want)
			}
		})
	}
}

func TestInquireMIDVA_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusInternalServerError, `{"error":"internal","error_code":"INTERNAL","message":"provider unavailable"}`)

	_, err := client.InquireMIDVA(context.Background(), &goholdstation.InquireMIDVARequest{VAAccount: "9704361234"})
	assertAPIError(t, err, http.StatusInternalServerError, "INTERNAL")
}
