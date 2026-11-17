package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestListFundLogs_Success_BareAndWrapped(t *testing.T) {
	t.Parallel()

	const payload = `{"success":true,"code":0,"error":"","logs":[{"id":42,"partner_fund_id":1,"change_type":1,"change_amount":"995000","fee":"5000","pre_balance":"0","post_balance":"995000","metadata":{"source":"fiat_deposit","reference_id":"dep_1","deposit_va_id":"dva_01HQ8X2M4K","account_number":"9001","sender_name":"NGUYEN VAN A","paid_at":"2026-10-29T03:14:22Z"},"caused_at":{"seconds":1700000000,"nanos":0}}],"total_count":1,"page":1,"page_size":20}`

	for name, body := range map[string]string{"bare": payload, "wrapped": `{"message":"success","data":` + payload + `}`} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client, api := newFakeAPI(t, http.StatusOK, body)

			resp, err := client.ListFundLogs(context.Background(), &goholdstation.ListFundLogsRequest{
				Token: "USDT", ChangeType: goholdstation.FundChangeTypeDeposit, Page: 1, PageSize: 20,
			})
			if err != nil {
				t.Fatalf("ListFundLogs: %v", err)
			}

			api.assertRequest(t, wantRequest{
				method: http.MethodGet, path: "/partners/funds/logs",
				query: "change_type=1&page=1&page_size=20&token=USDT", body: "", signed: true,
			})

			data := resp.Data
			if data.TotalCount != 1 || data.Page != 1 || data.PageSize != 20 || len(data.Logs) != 1 {
				t.Fatalf("data = %+v", data)
			}

			log := data.Logs[0]
			if log.ID != 42 || log.ChangeType != goholdstation.FundChangeTypeDeposit || log.ChangeAmount != "995000" || log.Fee != "5000" || log.PostBalance != "995000" {
				t.Errorf("log = %+v", log)
			}
		})
	}
}

func TestListFundLogs_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusBadRequest, badRequestBody)

	_, err := client.ListFundLogs(context.Background(), nil)
	assertAPIError(t, err, http.StatusBadRequest, "BAD_REQUEST")
}
