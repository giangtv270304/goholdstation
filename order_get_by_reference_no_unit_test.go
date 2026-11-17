package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestGetOrderByReferenceNo_Success_BothDocumentedForms(t *testing.T) {
	t.Parallel()

	bodies := map[string]string{
		"order":  `{"message":"success","data":{"order":` + depositOrderJSON + `,"payments":[{"id":"pay_1","amount":"1000000"}],"total_received":"1000000"}}`,
		"orders": `{"message":"success","data":{"orders":[` + depositOrderJSON + `],"total_count":1}}`,
	}

	for name, body := range bodies {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client, api := newFakeAPI(t, http.StatusOK, body)

			resp, err := client.GetOrderByReferenceNo(context.Background(), &goholdstation.GetOrderByReferenceNoRequest{ReferenceNo: "ref 1/a"})
			if err != nil {
				t.Fatalf("GetOrderByReferenceNo: %v", err)
			}

			api.assertRequest(t, wantRequest{method: http.MethodGet, path: "/partners/orders/reference-no/ref%201%2Fa", query: "", body: "", signed: false})

			data := resp.Data
			if data.Order == nil || data.Order.ReferenceNo != "ref-1" || data.Order.Amount != "1000000" {
				t.Fatalf("order = %+v", data.Order)
			}

			if len(data.Orders) != 1 || data.Orders[0].ID != data.Order.ID {
				t.Errorf("orders = %+v", data.Orders)
			}

			if name == "order" && (len(data.Payments) != 1 || data.TotalReceived != "1000000") {
				t.Errorf("payments = %+v total = %q", data.Payments, data.TotalReceived)
			}

			if name == "orders" && data.TotalCount != 1 {
				t.Errorf("total_count = %d, want 1", data.TotalCount)
			}
		})
	}
}

func TestGetOrderByReferenceNo_EmptyOrdersLeavesOrderNil(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusOK, `{"message":"success","data":{"orders":[],"total_count":0}}`)

	resp, err := client.GetOrderByReferenceNo(context.Background(), &goholdstation.GetOrderByReferenceNoRequest{ReferenceNo: "none"})
	if err != nil {
		t.Fatalf("GetOrderByReferenceNo: %v", err)
	}

	if resp.Data.Order != nil || len(resp.Data.Orders) != 0 {
		t.Errorf("data = %+v", resp.Data)
	}
}

func TestGetOrderByReferenceNo_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusNotFound, `{"error":"not found","error_code":"ORDER_NOT_FOUND","message":"order not found"}`)

	_, err := client.GetOrderByReferenceNo(context.Background(), &goholdstation.GetOrderByReferenceNoRequest{ReferenceNo: "missing"})
	assertAPIError(t, err, http.StatusNotFound, "ORDER_NOT_FOUND")
}
