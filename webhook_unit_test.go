package goholdstation_test

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

const orderWebhookBody = `{"id":"7f1c0f0e-3b1a-4c55-9d53-0a4a4f2f6f10","topic":"pay.order.status-updated","ts":"2026-10-29T03:14:22Z","payload":{"order_id":"ord_1","old_order_state":2,"new_order_state":3,"old_order_processing_state":13,"new_order_processing_state":14,"tx_hash":"0xabc","chain_id":56,"scan_url":"https://bscscan.com/tx/0xabc"}}`

func signWebhook(t *testing.T, body string) (signature, rawKey, spkiKey string) {
	t.Helper()

	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	der, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		t.Fatalf("MarshalPKIXPublicKey: %v", err)
	}

	return base64.StdEncoding.EncodeToString(ed25519.Sign(privateKey, []byte(body))),
		base64.StdEncoding.EncodeToString(publicKey),
		base64.StdEncoding.EncodeToString(der)
}

func TestVerifyWebhook(t *testing.T) {
	t.Parallel()

	signature, rawKey, spkiKey := signWebhook(t, orderWebhookBody)

	for name, key := range map[string]string{"raw 32-byte key": rawKey, "SPKI DER key": spkiKey} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if err := goholdstation.VerifyWebhookSignature([]byte(orderWebhookBody), signature, key); err != nil {
				t.Fatalf("valid signature rejected: %v", err)
			}

			changed := []byte(orderWebhookBody)
			changed[len(changed)/2] ^= 0x01

			err := goholdstation.VerifyWebhookSignature(changed, signature, key)
			if !errors.Is(err, goholdstation.ErrInvalidWebhookSignature) {
				t.Errorf("one changed byte: error = %v, want ErrInvalidWebhookSignature", err)
			}

			err = goholdstation.VerifyWebhookSignature(append([]byte(orderWebhookBody), ' '), signature, key)
			if !errors.Is(err, goholdstation.ErrInvalidWebhookSignature) {
				t.Errorf("trailing space: error = %v, want ErrInvalidWebhookSignature", err)
			}
		})
	}
}

func TestVerifyWebhook_RejectsBadInputs(t *testing.T) {
	t.Parallel()

	signature, rawKey, _ := signWebhook(t, orderWebhookBody)
	_, otherKey, _ := signWebhook(t, orderWebhookBody)
	body := []byte(orderWebhookBody)

	testCases := map[string]struct {
		signature string
		key       string
		want      error
	}{
		"empty signature":   {"", rawKey, goholdstation.ErrInvalidWebhookSignature},
		"not base64":        {"***", rawKey, goholdstation.ErrInvalidWebhookSignature},
		"short signature":   {"AAAA", rawKey, goholdstation.ErrInvalidWebhookSignature},
		"another key":       {signature, otherKey, goholdstation.ErrInvalidWebhookSignature},
		"empty key":         {signature, "", goholdstation.ErrInvalidChecksumKey},
		"key of wrong size": {signature, "AAAA", goholdstation.ErrInvalidChecksumKey},
	}

	for name, testCase := range testCases {
		if err := goholdstation.VerifyWebhookSignature(body, testCase.signature, testCase.key); !errors.Is(err, testCase.want) {
			t.Errorf("%s: error = %v, want %v", name, err, testCase.want)
		}
	}
}

func TestParseWebhookEvent_OrderStatusUpdated(t *testing.T) {
	t.Parallel()

	event, err := goholdstation.ParseWebhookEvent([]byte(orderWebhookBody))
	if err != nil {
		t.Fatalf("ParseWebhookEvent: %v", err)
	}

	if event.ID != "7f1c0f0e-3b1a-4c55-9d53-0a4a4f2f6f10" || event.Topic != goholdstation.WebhookTopicOrderStatusUpdated || event.TS != "2026-10-29T03:14:22Z" {
		t.Errorf("event = %+v", event)
	}

	payload, err := event.OrderStatusUpdated()
	if err != nil {
		t.Fatalf("OrderStatusUpdated: %v", err)
	}

	if payload.OrderID != "ord_1" || payload.OldOrderState != goholdstation.OrderStateProcessing || payload.NewOrderState != goholdstation.OrderStateCompleted {
		t.Errorf("payload = %+v", payload)
	}

	if payload.NewOrderProcessingState == nil || *payload.NewOrderProcessingState != goholdstation.OrderProcessingStateOnrampCryptoConfirmed {
		t.Errorf("new processing state = %v", payload.NewOrderProcessingState)
	}

	if payload.TxHash != "0xabc" || payload.ChainID == nil || *payload.ChainID != 56 || payload.ScanURL == "" {
		t.Errorf("payload = %+v", payload)
	}

	if _, err := event.FundBalanceChanged(); err == nil {
		t.Error("FundBalanceChanged accepted an order event")
	}
}

func TestParseWebhookEvent_RefundedWithoutProcessingStatesAndDeprecatedKeys(t *testing.T) {
	t.Parallel()

	const body = `{"id":"evt_2","type":"pay.order.status-updated","timestamp":"2026-10-29T03:14:22Z","payload":{"order_id":"ord_2","old_order_state":2,"new_order_state":6,"tx_hash":"0xdef","chain_id":56}}`

	event, err := goholdstation.ParseWebhookEvent([]byte(body))
	if err != nil {
		t.Fatalf("ParseWebhookEvent: %v", err)
	}

	if event.Topic != goholdstation.WebhookTopicOrderStatusUpdated || event.TS != "2026-10-29T03:14:22Z" {
		t.Errorf("event = %+v, want topic and ts from the deprecated keys", event)
	}

	payload, err := event.OrderStatusUpdated()
	if err != nil {
		t.Fatalf("OrderStatusUpdated: %v", err)
	}

	if payload.NewOrderState != goholdstation.OrderStateRefunded || payload.OldOrderProcessingState != nil || payload.NewOrderProcessingState != nil {
		t.Errorf("payload = %+v", payload)
	}
}

func TestParseWebhookEvent_FundBalanceChanged(t *testing.T) {
	t.Parallel()

	const body = `{"id":"evt_3","topic":"pay.partner.fund-balance-changed","ts":"2026-10-29T03:14:22Z","payload":{"partner_id":"partner-xyz","token":"USDT","pre_balance":"1000.500000000000000001","post_balance":"900.500000000000000001","change_amount":"-100","change_type":2,"fund_log_id":42}}`

	event, err := goholdstation.ParseWebhookEvent([]byte(body))
	if err != nil {
		t.Fatalf("ParseWebhookEvent: %v", err)
	}

	payload, err := event.FundBalanceChanged()
	if err != nil {
		t.Fatalf("FundBalanceChanged: %v", err)
	}

	want := goholdstation.FundBalanceChangedPayload{
		PartnerID: "partner-xyz", Token: "USDT",
		PreBalance: "1000.500000000000000001", PostBalance: "900.500000000000000001", ChangeAmount: "-100",
		ChangeType: goholdstation.FundChangeTypeWithdraw, FundLogID: 42,
	}

	if *payload != want {
		t.Errorf("payload = %+v, want %+v", *payload, want)
	}
}

func TestParseVAPaymentCallback(t *testing.T) {
	t.Parallel()

	const body = `{"va_account":"9001","va_account_name":"SHOP","va_account_addr":"HN","va_bank_name":"BIDV","transfer_content":"aGVsbG8=","amount":"150000","time_paid":"1700000000","transaction_id":"tx_1"}`

	callback, err := goholdstation.ParseVAPaymentCallback([]byte(body))
	if err != nil {
		t.Fatalf("ParseVAPaymentCallback: %v", err)
	}

	if callback.VAAccount != "9001" || callback.Amount != "150000" || callback.TimePaid != "1700000000" || callback.TransactionID != "tx_1" {
		t.Errorf("callback = %+v", callback)
	}

	if _, err := goholdstation.ParseVAPaymentCallback([]byte(`not json`)); err == nil {
		t.Error("ParseVAPaymentCallback accepted garbage")
	}

	if _, err := goholdstation.ParseWebhookEvent([]byte(`not json`)); err == nil {
		t.Error("ParseWebhookEvent accepted garbage")
	}
}
