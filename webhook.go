package goholdstation

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/giangtv270304/goholdstation/signer"
)

const (
	// WebhookSignatureHeader carries the signature of order and balance webhooks.
	WebhookSignatureHeader = "X-HSPay-Event-Signature"
	// WebhookTopicHeader carries the topic of order and balance webhooks.
	WebhookTopicHeader = "X-HSPay-Event-Topic"
	// VACallbackSignatureHeader carries the signature of the VA payment callback.
	VACallbackSignatureHeader = "X-HSPay-Response-Signature"

	WebhookTopicOrderStatusUpdated = "pay.order.status-updated"
	WebhookTopicFundBalanceChanged = "pay.partner.fund-balance-changed"
)

var (
	// ErrInvalidWebhookSignature is returned when the signature is missing, malformed or does not match the body.
	ErrInvalidWebhookSignature = errors.New("goholdstation: webhook signature is invalid")
	// ErrInvalidChecksumKey is returned when the Webhook Checksum Key cannot be decoded.
	ErrInvalidChecksumKey = errors.New("goholdstation: webhook checksum key is invalid")
)

// VerifyWebhookSignature checks the Ed25519 signature of a webhook or VA callback.
// body is the raw request body exactly as received; signature is the value of the signature header;
// checksumKey is the base64 Webhook Checksum Key (32 raw bytes or 44-byte SPKI DER).
// The API documents no timestamp: deduplicate deliveries on WebhookEvent.ID.
func VerifyWebhookSignature(body []byte, signature, checksumKey string) error {
	publicKey, err := signer.ParsePublicKey(checksumKey)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidChecksumKey, err)
	}

	if signature == "" {
		return fmt.Errorf("%w: signature header is empty", ErrInvalidWebhookSignature)
	}

	if err := signer.Verify(publicKey, body, signature); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidWebhookSignature, err)
	}

	return nil
}

// WebhookEvent is the envelope of order and balance webhooks.
type WebhookEvent struct {
	// ID is a UUID; use it to drop duplicate deliveries.
	ID    string `json:"id"`
	Topic string `json:"topic"`
	// TS is an RFC 3339 string.
	TS      string          `json:"ts"`
	Payload json.RawMessage `json:"payload"`
}

// UnmarshalJSON falls back to the deprecated "type" and "timestamp" keys when "topic" or "ts" is absent.
func (e *WebhookEvent) UnmarshalJSON(data []byte) error {
	type plain WebhookEvent

	shadow := struct {
		*plain

		Type      string `json:"type"`
		Timestamp string `json:"timestamp"`
	}{plain: (*plain)(e), Type: "", Timestamp: ""}

	if err := json.Unmarshal(data, &shadow); err != nil {
		return err
	}

	if e.Topic == "" {
		e.Topic = shadow.Type
	}

	if e.TS == "" {
		e.TS = shadow.Timestamp
	}

	return nil
}

// OrderStatusUpdatedPayload is the payload of WebhookTopicOrderStatusUpdated.
type OrderStatusUpdatedPayload struct {
	OrderID       string     `json:"order_id"`
	OldOrderState OrderState `json:"old_order_state"`
	NewOrderState OrderState `json:"new_order_state"`
	// The processing states are nil when the webhook leaves them out (the Refunded example does).
	OldOrderProcessingState *OrderProcessingState `json:"old_order_processing_state"`
	NewOrderProcessingState *OrderProcessingState `json:"new_order_processing_state"`
	// TxHash is the transfer to the user (onramp Completed) or the refund transfer (Refunded).
	TxHash  string `json:"tx_hash"`
	ChainID *int64 `json:"chain_id"`
	ScanURL string `json:"scan_url"`
}

// FundBalanceChangedPayload is the payload of WebhookTopicFundBalanceChanged.
// The three amounts are decimal strings. Read the details with GetFundLogByID.
type FundBalanceChangedPayload struct {
	PartnerID    string         `json:"partner_id"`
	Token        string         `json:"token"`
	PreBalance   string         `json:"pre_balance"`
	PostBalance  string         `json:"post_balance"`
	ChangeAmount string         `json:"change_amount"`
	ChangeType   FundChangeType `json:"change_type"`
	FundLogID    int32          `json:"fund_log_id"`
}

// VAPaymentCallback is the body of the VA payment callback. It has no envelope.
type VAPaymentCallback struct {
	VAAccount     string `json:"va_account"`
	VAAccountName string `json:"va_account_name"`
	VAAccountAddr string `json:"va_account_addr"`
	VABankName    string `json:"va_bank_name"`
	// TransferContent is base64.
	TransferContent string `json:"transfer_content"`
	// Amount is a decimal string in VND.
	Amount string `json:"amount"`
	// TimePaid is Unix seconds as a string.
	TimePaid      string `json:"time_paid"`
	TransactionID string `json:"transaction_id"`
}

// ParseWebhookEvent decodes the envelope of an order or balance webhook. Verify the signature first.
func ParseWebhookEvent(body []byte) (*WebhookEvent, error) {
	event := new(WebhookEvent)
	if err := json.Unmarshal(body, event); err != nil {
		return nil, fmt.Errorf("goholdstation: decode webhook event: %w", err)
	}

	return event, nil
}

// OrderStatusUpdated decodes the payload of an order status webhook. It fails on another topic.
func (e *WebhookEvent) OrderStatusUpdated() (*OrderStatusUpdatedPayload, error) {
	return decodeWebhookPayload[OrderStatusUpdatedPayload](e, WebhookTopicOrderStatusUpdated)
}

// FundBalanceChanged decodes the payload of a fund balance webhook. It fails on another topic.
func (e *WebhookEvent) FundBalanceChanged() (*FundBalanceChangedPayload, error) {
	return decodeWebhookPayload[FundBalanceChangedPayload](e, WebhookTopicFundBalanceChanged)
}

// ParseVAPaymentCallback decodes the body of a VA payment callback. Verify the signature first.
func ParseVAPaymentCallback(body []byte) (*VAPaymentCallback, error) {
	callback := new(VAPaymentCallback)
	if err := json.Unmarshal(body, callback); err != nil {
		return nil, fmt.Errorf("goholdstation: decode VA payment callback: %w", err)
	}

	return callback, nil
}

func decodeWebhookPayload[T any](event *WebhookEvent, topic string) (*T, error) {
	if event.Topic != topic {
		return nil, fmt.Errorf("goholdstation: webhook topic is %q, want %q", event.Topic, topic)
	}

	payload := new(T)
	if err := json.Unmarshal(event.Payload, payload); err != nil {
		return nil, fmt.Errorf("goholdstation: decode %s payload: %w", topic, err)
	}

	return payload, nil
}
