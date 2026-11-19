# goholdstation

Go client for the Holdstation Pay partner API.

Unofficial client, not affiliated with or endorsed by Holdstation.

## Install

```bash
go get github.com/giangtv270304/goholdstation
```

Requires Go 1.23+. No dependencies outside the standard library.

## First call

Ask the Holdstation team for a sandbox partner app key, then:

```go
client, err := goholdstation.NewClient(goholdstation.SandboxBaseURL, partnerAppKey)
if err != nil {
	log.Fatal(err)
}

resp, err := client.GetExchangeRate(ctx, nil)
if err != nil {
	log.Fatal(err)
}

fmt.Println(resp.Data.Rate.Buy, resp.Data.Rate.Sell)
```

A full program is in `example/quickstart`.

## Authentication

Every request carries the partner app key in the `Partner-App-Key` header. 30 of the 44 operations are also signed with your Ed25519 key (`X-HSPay-Signature`, `X-HSPay-Timestamp`); the client does it for you when you pass the key:

```go
client, err := goholdstation.NewClient(goholdstation.SandboxBaseURL, partnerAppKey,
	goholdstation.WithPrivateKey(ed25519PrivateKey))
```

- The app key and the private key are for servers only: never ship them in a browser or a mobile app.
- The private key is base64 (standard alphabet) of the 64-byte Ed25519 key, the 32-byte seed, or the 48-byte PKCS#8 DER. You generate the pair and give the base64 public key to the Holdstation team.
- A signed operation called on a client without `WithPrivateKey` returns `goholdstation.ErrMissingPrivateKey` before any network call.
- The server rejects a timestamp more than 60 seconds away from its own clock. Keep the host clock in sync.
- Signed operations: every funds, deposit account, MID, KYC, KYB and webhook-URL method, plus `DecodeQR`, `CreateWithdrawOrder`, `ConfirmWithdrawOrder` and `CreateCryptoPayoutOrder`.

| Environment | Base URL |
|---|---|
| Sandbox | `goholdstation.SandboxBaseURL` |
| Production | `goholdstation.ProductionBaseURL` |

The base URL must be `https` (plain `http` is accepted on a loopback host only, for tests).

## Flows

Runnable programs, configured from `.env.example`: `example/quickstart`, `example/onramp` (create, read and close a deposit order; `-keep` leaves it open), `example/prefund_withdraw` (quote; `-confirm` sends the payout), `example/webhook`.

### Onramp: the user pays VND and receives tokens

1. `CreateDepositOrder` returns the order; `Order.PayData` holds the bank account and QR code to pay.
2. Holdstation calls your `Callback` URL when the state changes (see Webhooks).
3. `GetOrderByID` or `GetOrderByReferenceNo` returns the current state. `CloseOrder` cancels an unpaid order.

```go
buffer := goholdstation.Decimal("0.005")

created, err := client.CreateDepositOrder(ctx, &goholdstation.CreateDepositOrderRequest{
	Amount: "1000000", Currency: "vnd", ChainID: 56, TokenAddress: tokenAddress,
	Recipient: userWallet, Callback: "https://yourapp.com/webhooks/holdstation",
	ReferenceNo: "order-1001", Buffer: &buffer,
})
```

### Prefund withdrawal: pay VND to a bank account from your prefunded balance

1. `CreateWithdrawOrder` returns a quote, held for up to 15 minutes. It moves no money.
2. `ConfirmWithdrawOrder` debits the fund and queues the payout.
3. The order webhook, or `GetOrderByID`, tells you when the payout is done.

`GetFundBalance`, `ListFundLogs` and `ConvertFund` manage the prefunded balances. `CreateCryptoPayoutOrder` pays stablecoins to a wallet.

### When the outcome of a money call is unknown

This applies to `CreateDepositOrder`, `CreateWithdrawalOrder`, `CreateWithdrawOrder`, `ConfirmWithdrawOrder`, `CreateCryptoPayoutOrder` and `ConvertFund`. The client never retries.

Only an `*APIError` with a 4xx status, or with a business code (status 200), tells you the call was refused. Any other error (timeout, network error, 5xx, `ErrInvalidResponse`) means the outcome is unknown. Then:

- look the order up with `GetOrderByID` or `GetOrderByReferenceNo`, or
- send the same request again with the same `IdempotencyKey` / `ReferenceNo`. Never use a new one.

`ConfirmWithdrawOrder` has no idempotency key: read the order first. `CreateDepositVA` is not idempotent either: every call opens a new account.

### Webhooks

Holdstation signs the raw body with Ed25519. Verify it with the Webhook Checksum Key before you parse anything:

```go
func handle(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		http.Error(w, "body too large", http.StatusRequestEntityTooLarge)
		return
	}

	signature := r.Header.Get(goholdstation.WebhookSignatureHeader)
	if err := goholdstation.VerifyWebhookSignature(body, signature, checksumKey); err != nil {
		http.Error(w, "bad signature", http.StatusUnauthorized)
		return
	}

	event, err := goholdstation.ParseWebhookEvent(body)
	if err != nil {
		http.Error(w, "bad body", http.StatusBadRequest)
		return
	}

	switch event.Topic {
	case goholdstation.WebhookTopicOrderStatusUpdated:
		payload, _ := event.OrderStatusUpdated()
		_ = payload
	case goholdstation.WebhookTopicFundBalanceChanged:
		payload, _ := event.FundBalanceChanged()
		_ = payload
	}

	w.WriteHeader(http.StatusOK)
}
```

- Pass the body exactly as received. Do not re-encode it.
- The API documents no timestamp or replay window: drop duplicates on `event.ID`.
- Answer HTTP 200 within 30 seconds; Holdstation retries up to 5 times.
- The VA payment callback has no envelope: verify with the `X-HSPay-Response-Signature` header (`goholdstation.VACallbackSignatureHeader`) and decode with `ParseVAPaymentCallback`.
- The checksum key is accepted as base64 of the raw 32-byte key or of the 44-byte SPKI DER.

## Methods

| Method | What it does | Signed |
|---|---|---|
| `ListBanks` | One page of banks | no |
| `VerifyBankAccount` | Holder name of a bank account | no |
| `DecodeQR` | Decodes a bank QR code | yes |
| `GetTransactionThreshold` | Transaction threshold configuration | no |
| `GetExchangeRate` | USD/VND rate | no |
| `ListTokens` | One page of tokens | no |
| `CreateWholesaleApplication` | Submits a wholesale application | no |
| `EstimateOrderOutcome` | Quotes an order without creating it | no |
| `CreateDepositOrder` | Creates an onramp order | no |
| `CreateWithdrawalOrder` | Creates an offramp order | no |
| `ListOrders` | One page of orders | no |
| `GetOrderByID` | One order | no |
| `GetOrderByReferenceNo` | The order with your reference number | no |
| `GetOrderCustomerInfo` | Who paid an onramp order | no |
| `CloseOrder` | Closes a pending order | no |
| `CreateWithdrawOrder` | Prefund withdrawal, step 1 (quote) | yes |
| `ConfirmWithdrawOrder` | Prefund withdrawal, step 2 (debit and pay) | yes |
| `CreateCryptoPayoutOrder` | Pays stablecoins to a wallet | yes |
| `ListFunds` | Every prefunded balance | yes |
| `GetFundBalance` | Prefunded balance of one token | yes |
| `ConvertFund` | Converts a balance between stablecoins | yes |
| `ListFundDepositAddresses` | Crypto deposit addresses | yes |
| `ListFundLogs` | One page of balance changes | yes |
| `GetFundLogByID` | One balance change | yes |
| `CreateDepositVA` | Opens a VND deposit account | yes |
| `ListDepositVAs` | VND deposit accounts | yes |
| `GetDepositVAByID` | One VND deposit account | yes |
| `CloseDepositVA` | Closes a VND deposit account (irreversible) | yes |
| `ListDepositVATransactions` | Transfers received on a deposit account | yes |
| `GetMIDBalance` | Merchant (MID) balance | yes |
| `CreateMIDVA` | Creates a merchant virtual account | yes |
| `ListMIDVAs` | Merchant virtual accounts | yes |
| `GetMIDVAByAccount` | One merchant virtual account | yes |
| `RevokeMIDVA` | Revokes a merchant virtual account | yes |
| `InquireMIDVA` | Payments received on a merchant virtual account | yes |
| `GetWebhookURL` | URL of the fund balance webhook | yes |
| `SetWebhookURL` | Sets or clears that URL | yes |
| `SubmitKYB` | Submits a business for verification | yes |
| `ListKYBs` | One page of KYB records | yes |
| `GetKYBByID` | One KYB record | yes |
| `CreateKYC` | Stores a customer identity record | yes |
| `UpdateKYC` | Updates a customer identity record | yes |
| `GetKYCByID` | One customer identity record | yes |
| `ListKYCs` | One page of customer identity records | yes |

Every method returns `*Response[T]`: `Data` is the payload, `Message` is the envelope message (empty when the API answers without the `{message, data}` envelope).

Amounts, balances and rates are decimal strings (`"1000.50"`). Where the API uses a JSON number (`buffer`, `total_fee_vnd`, fees, token rates), the Go type is `goholdstation.Decimal`: it keeps the digits as text and is written back as a JSON number. Do not parse either into a float.

## Errors

Every response with status 400 or above is a `*goholdstation.APIError`. So is an HTTP 200 response that reports `success: false`: `StatusCode` is 200 and `Code` is the business code (`"3"` is insufficient balance on a withdrawal).

```go
var apiErr *goholdstation.APIError
if errors.As(err, &apiErr) {
	log.Printf("%s (status %d, code %s)", apiErr.Message, apiErr.StatusCode, apiErr.Code)
}

if goholdstation.IsNotFound(err) {
	// 404
}
```

- The API has no request id. Keep `apiErr.Body` (the raw body) when you report a problem.
- `apiErr.Details` lists the validation failures of a 422 (`Code`, `Target`).
- `IsNotFound`, `IsUnauthorized`, `IsConflict`, `IsRateLimited` check the status.
- `GetOrderCustomerInfo` and `DecodeQR` return their "not available" and "decode failed" results in the response, not as errors.
- A timeout is `errors.Is(err, context.DeadlineExceeded)`. A response the client cannot use (3xx, a 2xx body that does not decode, a body over 10 MiB) is `errors.Is(err, goholdstation.ErrInvalidResponse)`.

## Known limitations

The vendor documentation contradicts itself in places, and nothing was checked against a live server yet.

- Where two response forms are documented, the client decodes both: bare or `{message, data}` envelope on the funds, webhook-URL and QR endpoints; `order` or `orders` on `GetOrderByReferenceNo`; `kyc` or the flat record on KYC; paged object or plain array on `ListDepositVAs`, `ListKYCs`, `ListTokens`; `{seconds, nanos}` or RFC 3339 timestamps; camelCase or snake_case on `RevokeMIDVA` and `InquireMIDVA`; both field sets on `GetMIDBalance`.
- The business codes of `CloseOrder`, `ConvertFund` and most `CreateCryptoPayoutOrder` failures are not documented. The docs also show some withdrawal failures as HTTP 400/404 instead of HTTP 200 with a code; handle both.
- `currency` on deposit and withdrawal orders is documented as `vnd | token` and as `vnd | usd`. `buffer`, `reference_no` and `kyc_id` are documented as required and as optional; the client leaves them out when empty.
- `CreateDepositVA` sends `va_name` when set and no body otherwise (both are documented).
- `Order.Body`, `FundLog.Metadata` and `Config.Value` are raw JSON: their wire form is not settled. `FundLogMetadata` fits the plain-JSON form.
- `OrderType` values for prefund withdraw and crypto payout orders, order state 6 (Refunded) and the "waiting for approval" payout states are not fully documented. Unknown values decode without error.
- Not in this SDK (their paths are unconfirmed): rotating the signing key (`verify-key`), the VA callback URL, maintenance windows.
- `CreateWholesaleApplication` may not be callable with a partner key.
- `SubmitKYB` and `CreateKYC` can take up to 120 seconds: use an HTTP client with a longer timeout.

## Configuration

```go
client, err := goholdstation.NewClient(baseURL, partnerAppKey,
	goholdstation.WithPrivateKey(ed25519PrivateKey),
	goholdstation.WithHTTPClient(&http.Client{
		Timeout:       120 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}),
	goholdstation.WithLogger(slog.Default()),
)
```

The default timeout is 30 seconds. The default client does not follow redirects; a custom client should not either, because a redirect would carry the app key to another host. Debug logs carry the method, path and status, never query strings, headers, bodies or keys. Printing a `*Client` never shows the keys.

## Tests

```bash
go test ./...
```

Sandbox tests are skipped unless the variables in `.env.example` are set in the environment:

```bash
go test -run TestSandbox_ ./...
```

They never run against production: the test helper stops if `HOLDSTATION_BASE_URL` is the production URL. Each run leaves one Closed onramp order with a reference number starting with `sdk-sandbox-test-`. The first signed run also creates the KYC record `sdk-sandbox-test-kyc`; KYC records cannot be deleted.

## Versioning

Semantic versioning. The version is `goholdstation.Version`. Breaking changes are listed in the release notes of each major version.
