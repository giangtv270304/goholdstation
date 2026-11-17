package goholdstation_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestGetFundLogByID_Success_BareAndWrapped(t *testing.T) {
	t.Parallel()

	const payload = `{"success":true,"code":0,"error":"","log":{"id":42,"partner_fund_id":1,"change_type":1,"change_amount":"100.5","fee":"0","pre_balance":"0","post_balance":"100.5","metadata":{"source":"crypto_deposit","reference_id":"dep_2","chain_id":56,"tx_hash":"0xabc"},"caused_at":{"seconds":1700000000,"nanos":7}}}`

	for name, body := range map[string]string{"bare": payload, "wrapped": `{"message":"success","data":` + payload + `}`} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client, api := newFakeAPI(t, http.StatusOK, body)

			resp, err := client.GetFundLogByID(context.Background(), &goholdstation.GetFundLogByIDRequest{FundLogID: 42})
			if err != nil {
				t.Fatalf("GetFundLogByID: %v", err)
			}

			api.assertRequest(t, wantRequest{method: http.MethodGet, path: "/partners/funds/logs/42", query: "", body: "", signed: true})

			log := resp.Data.Log
			if log == nil || log.ID != 42 || log.PartnerFundID != 1 || log.ChangeAmount != "100.5" || log.CausedAt.Nanos != 7 {
				t.Fatalf("log = %+v", log)
			}

			var metadata goholdstation.FundLogMetadata
			if err := json.Unmarshal(log.Metadata, &metadata); err != nil {
				t.Fatalf("decode metadata: %v", err)
			}

			if metadata.Source != "crypto_deposit" || metadata.ReferenceID != "dep_2" || metadata.ChainID != 56 || metadata.TxHash != "0xabc" {
				t.Errorf("metadata = %+v", metadata)
			}
		})
	}
}

func TestGetFundLogByID_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusNotFound, notFoundBody)

	_, err := client.GetFundLogByID(context.Background(), &goholdstation.GetFundLogByIDRequest{FundLogID: 9})
	if !goholdstation.IsNotFound(err) {
		t.Errorf("error = %v, want IsNotFound", err)
	}

	apiErr := requireAPIError(t, err, http.StatusNotFound, "NOT_FOUND")
	if apiErr.Message != "No fund log with this ID" {
		t.Errorf("message = %q", apiErr.Message)
	}
}
