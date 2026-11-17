package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

const tokenJSON = `{"id":"0d3b1c4e-8f6a-4a57-9a43-2f0f5a1f7b11","chain_id":56,"address":"0x55d398326f99059ff775485246999027b3197955","name":"Tether USD","symbol":"USDT","name_override":"","decimals":18,"price":"1.0001","price_change_24h":"-0.01","image":"https://img.example/usdt.png","active":true,"withdrawal_contract":"0xabc","rate":{"buy":25850.5,"sell":25750},"created_at":{"seconds":1700000000,"nanos":0},"updated_at":{"seconds":1700000000,"nanos":0}}`

func TestListTokens_Success_EnvelopeAndBareArray(t *testing.T) {
	t.Parallel()

	bodies := map[string]string{
		"envelope":   `{"message":"success","data":{"tokens":[` + tokenJSON + `],"page":1,"page_size":50,"total_count":1}}`,
		"bare array": `[` + tokenJSON + `]`,
	}

	for name, body := range bodies {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client, api := newFakeAPI(t, http.StatusOK, body)
			active := false

			resp, err := client.ListTokens(context.Background(), &goholdstation.ListTokensRequest{
				ChainID: 56, Active: &active, Page: 1, PageSize: 50, Symbols: []string{"USDT", "USDC"}, NoLimit: true, WithRate: true,
			})
			if err != nil {
				t.Fatalf("ListTokens: %v", err)
			}

			api.assertRequest(t, wantRequest{
				method: http.MethodGet, path: "/api/v1/tokens",
				query: "active=false&chain_id=56&no_limit=true&page=1&page_size=50&symbols=USDT%2CUSDC&with_rate=true",
				body:  "", signed: false,
			})

			if len(resp.Data.Tokens) != 1 {
				t.Fatalf("tokens = %+v", resp.Data.Tokens)
			}

			token := resp.Data.Tokens[0]
			if token.ID != "0d3b1c4e-8f6a-4a57-9a43-2f0f5a1f7b11" || token.ChainID != 56 || token.Symbol != "USDT" || token.Decimals != 18 || !token.Active {
				t.Errorf("token = %+v", token)
			}

			if token.Price != "1.0001" || token.PriceChange24h != "-0.01" || token.Rate == nil || token.Rate.Buy != "25850.5" || token.Rate.Sell != "25750" {
				t.Errorf("token money fields = %+v rate %+v", token, token.Rate)
			}
		})
	}
}

func TestListTokens_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusBadRequest, badRequestBody)

	_, err := client.ListTokens(context.Background(), nil)
	assertAPIError(t, err, http.StatusBadRequest, "BAD_REQUEST")
}
