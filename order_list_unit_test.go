package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestListOrders_Success(t *testing.T) {
	t.Parallel()

	client, api := newFakeAPI(t, http.StatusOK, `{"message":"success","data":{"orders":[{"id":"ord_1","state":3,"processing_state":14,"amount":"1000000","kyc_id":"kyc-1","approval_reason":2,"body":{"k":"v"},"reference":{"hs_rate":25850.75,"usdt_rate":25800,"token_rate":1.0001,"partner_rate":25900,"spread":0.002,"buffer_rate":0.005,"buffered_rate_usdt":25929,"buffered_rate_partner":26029.5,"addition_share_rate":0.1,"partner_override_revenue":12.5,"provider_fee":1100,"token_symbol":"USDT","sim_webhook_id":"sim_1"}}],"total_count":1}}`)

	resp, err := client.ListOrders(context.Background(), &goholdstation.ListOrdersRequest{
		Types:            []goholdstation.OrderType{goholdstation.OrderTypeOnramp, goholdstation.OrderTypeOfframp},
		States:           []goholdstation.OrderState{goholdstation.OrderStateProcessing, goholdstation.OrderStateCompleted},
		ProcessingStates: []goholdstation.OrderProcessingState{goholdstation.OrderProcessingStateOnrampFiatPending},
		Page:             2, PageSize: 25, SortBy: "created_at", SortDesc: true,
		CreatedAtFrom: 1700000000, CreatedAtTo: 1700003600, ExpiredAtFrom: 1700000001, ExpiredAtTo: 1700003601,
		Recipient: "0xabc", ReferenceNos: []string{"ref-1", "ref-2"},
	})
	if err != nil {
		t.Fatalf("ListOrders: %v", err)
	}

	api.assertRequest(t, wantRequest{
		method: http.MethodGet, path: "/partners/orders",
		query:  "created_at_from=1700000000&created_at_to=1700003600&expired_at_from=1700000001&expired_at_to=1700003601&page=2&page_size=25&processing_state=10&recipient=0xabc&reference_nos=ref-1%2Cref-2&sort_by=created_at&sort_desc=true&state=2%2C3&type=1%2C2",
		body:   "",
		signed: false,
	})

	if resp.Data.TotalCount != 1 || len(resp.Data.Orders) != 1 {
		t.Fatalf("data = %+v", resp.Data)
	}

	order := resp.Data.Orders[0]
	if order.ID != "ord_1" || order.State != goholdstation.OrderStateCompleted || order.KYCID != "kyc-1" || order.ApprovalReason != 2 || string(order.Body) != `{"k":"v"}` {
		t.Errorf("order = %+v", order)
	}

	reference := order.Reference
	if reference == nil || reference.HSRate != "25850.75" || reference.TokenRate != "1.0001" || reference.BufferedRatePartner != "26029.5" || reference.ProviderFee != "1100" || reference.TokenSymbol != "USDT" {
		t.Errorf("reference = %+v", reference)
	}
}

func TestListOrders_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusUnauthorized, unauthorizedBody)

	_, err := client.ListOrders(context.Background(), nil)
	assertAPIError(t, err, http.StatusUnauthorized, "UNAUTHORIZED")
}
