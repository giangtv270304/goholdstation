package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestResponse_SuccessFalseOnHTTP200IsAnAPIError(t *testing.T) {
	t.Parallel()

	bodies := map[string]string{
		"wrapped": `{"message":"success","data":{"success":false,"code":3,"error":"insufficient balance","order":null}}`,
		"bare":    `{"success":false,"code":3,"error":"insufficient balance"}`,
	}

	for name, body := range bodies {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client, _ := newFakeAPI(t, http.StatusOK, body)

			resp, err := client.ConfirmWithdrawOrder(context.Background(), &goholdstation.ConfirmWithdrawOrderRequest{OrderID: "ord_1"})
			if resp != nil {
				t.Errorf("resp = %+v, want nil", resp)
			}

			apiErr := requireAPIError(t, err, http.StatusOK, "3")
			if apiErr.Message != "insufficient balance" || string(apiErr.Body) != body {
				t.Errorf("api error = %+v", apiErr)
			}
		})
	}
}

func TestResponse_SuccessFalseWithoutMessage(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusOK, `{"message":"success","data":{"success":false}}`)

	_, err := client.CloseOrder(context.Background(), &goholdstation.CloseOrderRequest{OrderID: "ord_1"})

	apiErr := requireAPIError(t, err, http.StatusOK, "")
	if apiErr.Message == "" {
		t.Error("message is empty")
	}
}

func TestResponse_CodeAndErrorWithoutSuccessIsAnAPIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusOK, `{"data":{"code":1,"error":"webhook_url must be a valid public HTTPS URL"}}`)

	_, err := client.SetWebhookURL(context.Background(), &goholdstation.SetWebhookURLRequest{WebhookURL: "http://x"})

	apiErr := requireAPIError(t, err, http.StatusOK, "1")
	if apiErr.Message != "webhook_url must be a valid public HTTPS URL" {
		t.Errorf("message = %q", apiErr.Message)
	}
}

func TestResponse_SuccessTrueWithCodeZeroIsNotAnError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusOK, `{"message":"success","data":{"success":true,"code":0,"error":""}}`)

	resp, err := client.CloseOrder(context.Background(), &goholdstation.CloseOrderRequest{OrderID: "ord_1"})
	if err != nil {
		t.Fatalf("CloseOrder: %v", err)
	}

	if !resp.Data.Success || resp.Message != "success" {
		t.Errorf("resp = %+v", resp)
	}
}

func TestResponse_UnknownEnumValuesAndExtraFieldsDecode(t *testing.T) {
	t.Parallel()

	const body = `{"message":"success","extra_top":1,"data":{"order":{"id":"ord_1","order_type":77,"state":7,"processing_state":999,"brand_new_field":{"a":[1,2]},"fees":{"service_fee":5000,"min_fee":0.01,"other":true}},"new_section":[]}}`

	client, _ := newFakeAPI(t, http.StatusOK, body)

	resp, err := client.GetOrderByID(context.Background(), &goholdstation.GetOrderByIDRequest{OrderID: "ord_1"})
	if err != nil {
		t.Fatalf("GetOrderByID: %v", err)
	}

	order := resp.Data.Order
	if order.OrderType != 77 || order.State != 7 || order.ProcessingState != 999 {
		t.Errorf("order = %+v, want the unknown enum values kept", order)
	}

	if order.Fees.ServiceFee != "5000" || order.Fees.MinFee != "0.01" {
		t.Errorf("fees = %+v", order.Fees)
	}
}
