package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

func TestGetOrderCustomerInfo_Success(t *testing.T) {
	t.Parallel()

	client, api := newFakeAPI(t, http.StatusOK, `{"message":"success","data":{"success":true,"code":0,"error":"","customer_info":{"customer_account":"****4567","customer_bank_code":"970436","customer_name":"NGUYEN VAN A"}}}`)

	resp, err := client.GetOrderCustomerInfo(context.Background(), &goholdstation.GetOrderCustomerInfoRequest{OrderID: "ord_1"})
	if err != nil {
		t.Fatalf("GetOrderCustomerInfo: %v", err)
	}

	api.assertRequest(t, wantRequest{method: http.MethodGet, path: "/partners/orders/ord_1/customer-info", query: "", body: "", signed: false})

	info := resp.Data.CustomerInfo
	if !resp.Data.Success || resp.Data.Code != goholdstation.CustomerInfoCodeSuccess || info == nil || info.CustomerAccount != "****4567" || info.CustomerBankCode != "970436" || info.CustomerName != "NGUYEN VAN A" {
		t.Errorf("data = %+v info = %+v", resp.Data, info)
	}
}

func TestGetOrderCustomerInfo_NotAvailableIsReturnedNotAnError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusOK, `{"message":"success","data":{"success":false,"code":3,"error":"customer info not yet available","customer_info":null}}`)

	resp, err := client.GetOrderCustomerInfo(context.Background(), &goholdstation.GetOrderCustomerInfoRequest{OrderID: "ord_1"})
	if err != nil {
		t.Fatalf("GetOrderCustomerInfo: %v", err)
	}

	data := resp.Data
	if data.Success || data.Code != goholdstation.CustomerInfoCodeNotAvailable || data.Error != "customer info not yet available" || data.CustomerInfo != nil {
		t.Errorf("data = %+v", data)
	}
}

func TestGetOrderCustomerInfo_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusNotFound, `{"error":"not found","error_code":"ORDER_NOT_FOUND","message":"order not found"}`)

	_, err := client.GetOrderCustomerInfo(context.Background(), &goholdstation.GetOrderCustomerInfoRequest{OrderID: "missing"})
	assertAPIError(t, err, http.StatusNotFound, "ORDER_NOT_FOUND")
}
