package goholdstation_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/giangtv270304/goholdstation"
	"github.com/giangtv270304/goholdstation/signer"
)

func TestNewClient_RequiredArguments(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		baseURL string
		appKey  string
		want    error
	}{
		{"empty base url", "", "key", goholdstation.ErrEmptyBaseURL},
		{"empty app key", "https://x.example", "", goholdstation.ErrEmptyAppKey},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			_, err := goholdstation.NewClient(testCase.baseURL, testCase.appKey)
			if !errors.Is(err, testCase.want) {
				t.Errorf("error = %v, want %v", err, testCase.want)
			}
		})
	}
}

func TestNewClient_InvalidPrivateKeyNeverLeaksIt(t *testing.T) {
	t.Parallel()

	const badKey = "bm90LWEtcmVhbC1rZXk="

	_, err := goholdstation.NewClient("https://x.example", "key", goholdstation.WithPrivateKey(badKey))
	if !errors.Is(err, signer.ErrInvalidPrivateKey) {
		t.Fatalf("error = %v, want signer.ErrInvalidPrivateKey", err)
	}

	if strings.Contains(err.Error(), badKey) {
		t.Errorf("error message contains the key: %v", err)
	}
}

func TestClient_PrintingNeverLeaksKeys(t *testing.T) {
	t.Parallel()

	_, privateKey := newTestKey(t)

	client, err := goholdstation.NewClient("https://x.example", testAppKey, goholdstation.WithPrivateKey(privateKey))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	for _, format := range []string{"%v", "%+v", "%#v", "%s"} {
		printed := fmt.Sprintf(format, client)
		if strings.Contains(printed, testAppKey) || strings.Contains(printed, privateKey) || !strings.Contains(printed, "x.example") {
			t.Errorf("%s prints %q", format, printed)
		}
	}
}

func TestClient_UnsignedOperation_SendsKeyAndUserAgentWithoutSignature(t *testing.T) {
	t.Parallel()

	client, api := newFakeAPI(t, http.StatusOK, `{"message":"success","data":{"rate":{"buy":"25850","sell":"25750"}}}`)

	if _, err := client.GetExchangeRate(context.Background(), nil); err != nil {
		t.Fatalf("GetExchangeRate: %v", err)
	}

	api.assertRequest(t, wantRequest{method: http.MethodGet, path: "/api/v1/exchange/rate", query: "", body: "", signed: false})

	if got := api.header.Get("X-HSPay-Signature") + api.header.Get("X-HSPay-Timestamp"); got != "" {
		t.Errorf("signature headers = %q, want none", got)
	}
}

func TestClient_SignedOperation_UsesInjectedClockAndSortedQuery(t *testing.T) {
	t.Parallel()

	fixed := time.Unix(1700000000, 0)
	client, api := newFakeAPI(t, http.StatusOK, `{"success":true,"code":0,"error":"","logs":[],"total_count":0,"page":1,"page_size":20}`,
		goholdstation.WithClock(func() time.Time { return fixed }))

	_, err := client.ListFundLogs(context.Background(), &goholdstation.ListFundLogsRequest{
		Token: "US DT&x", ChangeType: goholdstation.FundChangeTypeDeposit, Page: 1, PageSize: 20,
	})
	if err != nil {
		t.Fatalf("ListFundLogs: %v", err)
	}

	api.assertRequest(t, wantRequest{
		method: http.MethodGet,
		path:   "/partners/funds/logs",
		query:  "change_type=1&page=1&page_size=20&token=US+DT%26x",
		body:   "",
		signed: true,
	})

	if got := api.header.Get("X-HSPay-Timestamp"); got != "1700000000" {
		t.Errorf("X-HSPay-Timestamp = %q, want 1700000000", got)
	}
}

func TestClient_SignedOperationWithoutPrivateKey_FailsBeforeNetwork(t *testing.T) {
	t.Parallel()

	var calls int

	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { calls++ }))
	t.Cleanup(server.Close)

	for _, opts := range [][]goholdstation.Option{nil, {goholdstation.WithPrivateKey("")}} {
		client, err := goholdstation.NewClient(server.URL, testAppKey, opts...)
		if err != nil {
			t.Fatalf("NewClient: %v", err)
		}

		_, err = client.ListFunds(context.Background(), nil)
		if !errors.Is(err, goholdstation.ErrMissingPrivateKey) {
			t.Errorf("error = %v, want ErrMissingPrivateKey", err)
		}
	}

	if calls != 0 {
		t.Errorf("server received %d requests, want 0", calls)
	}
}

func TestClient_BrokenSuccessBody(t *testing.T) {
	t.Parallel()

	for name, body := range map[string]string{
		"truncated json": `{"message":"success","data":{"rate":`,
		"html":           `<html>ok</html>`,
		"empty":          ``,
		"null":           `null`,
		"wrong type":     `{"message":"success","data":{"rate":"not an object"}}`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client, _ := newFakeAPI(t, http.StatusOK, body)

			resp, err := client.GetExchangeRate(context.Background(), nil)
			if err == nil || resp != nil {
				t.Fatalf("resp = %v, error = %v, want a decode error", resp, err)
			}

			var apiErr *goholdstation.APIError
			if errors.As(err, &apiErr) || !errors.Is(err, goholdstation.ErrInvalidResponse) {
				t.Errorf("error = %v, want ErrInvalidResponse, not *APIError", err)
			}

			if !strings.Contains(err.Error(), "decode response of GET /api/v1/exchange/rate") {
				t.Errorf("error = %v, want it to name the call", err)
			}
		})
	}
}

type failingHTTPClient struct{ err error }

func (f failingHTTPClient) Do(*http.Request) (*http.Response, error) { return nil, f.err }

func TestClient_NetworkErrorIsWrapped(t *testing.T) {
	t.Parallel()

	errBoom := errors.New("connection refused")

	client, err := goholdstation.NewClient("https://x.example", testAppKey,
		goholdstation.WithHTTPClient(failingHTTPClient{err: errBoom}))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	_, err = client.GetExchangeRate(context.Background(), nil)
	if !errors.Is(err, errBoom) {
		t.Errorf("error = %v, want it to wrap %v", err, errBoom)
	}
}

func TestClient_Timeout(t *testing.T) {
	t.Parallel()

	release := make(chan struct{})

	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { <-release }))
	t.Cleanup(server.Close)
	t.Cleanup(func() { close(release) })

	client, err := goholdstation.NewClient(server.URL, testAppKey)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err = client.GetExchangeRate(ctx, nil)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("error = %v, want context.DeadlineExceeded", err)
	}
}

type recordingLogger struct {
	mu    sync.Mutex
	lines []string
}

func (l *recordingLogger) Debug(msg string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.lines = append(l.lines, msg+" "+fmt.Sprint(args...))
}

func TestClient_DebugLogHasNoHeadersBodyOrKeys(t *testing.T) {
	t.Parallel()

	logger := &recordingLogger{}
	client, api := newFakeAPI(t, http.StatusOK, `{"webhook_url":"https://response-secret.example/hook"}`, goholdstation.WithLogger(logger))

	_, err := client.SetWebhookURL(context.Background(), &goholdstation.SetWebhookURLRequest{WebhookURL: "https://request-secret.example/hook"})
	if err != nil {
		t.Fatalf("SetWebhookURL: %v", err)
	}

	if _, err := client.GetFundBalance(context.Background(), &goholdstation.GetFundBalanceRequest{Token: "query-secret"}); err != nil {
		t.Fatalf("GetFundBalance: %v", err)
	}

	logger.mu.Lock()
	logged := strings.Join(logger.lines, "\n")
	logger.mu.Unlock()

	if !strings.Contains(logged, "POST") || !strings.Contains(logged, "/partners/webhook-url") || !strings.Contains(logged, "200") {
		t.Errorf("log = %q, want method, URL and status", logged)
	}

	forbidden := []string{
		testAppKey,
		api.header.Get("X-HSPay-Signature"),
		"X-HSPay", "Partner-App-Key",
		"request-secret", "response-secret", "query-secret", "?",
	}

	for _, secret := range forbidden {
		if secret == "" {
			t.Fatal("test setup: empty forbidden value")
		}

		if strings.Contains(logged, secret) {
			t.Errorf("log contains %q: %s", secret, logged)
		}
	}
}

func TestClient_EmptyRequiredParams_FailBeforeNetwork(t *testing.T) {
	t.Parallel()

	client, api := newFakeAPI(t, http.StatusOK, `{}`)
	ctx := context.Background()

	calls := map[string]func() error{
		"GetOrderByID": func() error {
			_, err := client.GetOrderByID(ctx, &goholdstation.GetOrderByIDRequest{})
			return err
		},
		"GetOrderByReferenceNo": func() error {
			_, err := client.GetOrderByReferenceNo(ctx, &goholdstation.GetOrderByReferenceNoRequest{})
			return err
		},
		"GetOrderCustomerInfo": func() error {
			_, err := client.GetOrderCustomerInfo(ctx, &goholdstation.GetOrderCustomerInfoRequest{})
			return err
		},
		"CloseOrder": func() error {
			_, err := client.CloseOrder(ctx, &goholdstation.CloseOrderRequest{})
			return err
		},
		"ConfirmWithdrawOrder": func() error {
			_, err := client.ConfirmWithdrawOrder(ctx, &goholdstation.ConfirmWithdrawOrderRequest{})
			return err
		},
		"GetFundBalance": func() error {
			_, err := client.GetFundBalance(ctx, &goholdstation.GetFundBalanceRequest{})
			return err
		},
		"GetFundLogByID": func() error {
			_, err := client.GetFundLogByID(ctx, &goholdstation.GetFundLogByIDRequest{})
			return err
		},
		"GetDepositVAByID": func() error {
			_, err := client.GetDepositVAByID(ctx, &goholdstation.GetDepositVAByIDRequest{})
			return err
		},
		"CloseDepositVA": func() error {
			_, err := client.CloseDepositVA(ctx, &goholdstation.CloseDepositVARequest{})
			return err
		},
		"ListDepositVATransactions": func() error {
			_, err := client.ListDepositVATransactions(ctx, &goholdstation.ListDepositVATransactionsRequest{})
			return err
		},
		"GetMIDVAByAccount": func() error {
			_, err := client.GetMIDVAByAccount(ctx, &goholdstation.GetMIDVAByAccountRequest{})
			return err
		},
		"RevokeMIDVA": func() error {
			_, err := client.RevokeMIDVA(ctx, &goholdstation.RevokeMIDVARequest{})
			return err
		},
		"InquireMIDVA": func() error {
			_, err := client.InquireMIDVA(ctx, &goholdstation.InquireMIDVARequest{})
			return err
		},
		"GetKYBByID": func() error {
			_, err := client.GetKYBByID(ctx, &goholdstation.GetKYBByIDRequest{})
			return err
		},
		"CreateKYC": func() error {
			_, err := client.CreateKYC(ctx, &goholdstation.CreateKYCRequest{})
			return err
		},
		"UpdateKYC": func() error {
			_, err := client.UpdateKYC(ctx, &goholdstation.UpdateKYCRequest{})
			return err
		},
		"GetKYCByID": func() error {
			_, err := client.GetKYCByID(ctx, &goholdstation.GetKYCByIDRequest{})
			return err
		},
	}

	for name, run := range calls {
		if err := run(); !errors.Is(err, goholdstation.ErrMissingField) {
			t.Errorf("%s: error = %v, want ErrMissingField", name, err)
		}
	}

	assertMissingField(t, fmt.Errorf("%w: all", goholdstation.ErrMissingField), api)
}

func TestClient_NilRequest(t *testing.T) {
	t.Parallel()

	client, api := newFakeAPI(t, http.StatusOK, `{}`)
	ctx := context.Background()

	errs := map[string]error{}
	_, errs["VerifyBankAccount"] = client.VerifyBankAccount(ctx, nil)
	_, errs["DecodeQR"] = client.DecodeQR(ctx, nil)
	_, errs["CreateWholesaleApplication"] = client.CreateWholesaleApplication(ctx, nil)
	_, errs["EstimateOrderOutcome"] = client.EstimateOrderOutcome(ctx, nil)
	_, errs["CreateDepositOrder"] = client.CreateDepositOrder(ctx, nil)
	_, errs["CreateWithdrawalOrder"] = client.CreateWithdrawalOrder(ctx, nil)
	_, errs["GetOrderByID"] = client.GetOrderByID(ctx, nil)
	_, errs["CreateWithdrawOrder"] = client.CreateWithdrawOrder(ctx, nil)
	_, errs["ConfirmWithdrawOrder"] = client.ConfirmWithdrawOrder(ctx, nil)
	_, errs["CreateCryptoPayoutOrder"] = client.CreateCryptoPayoutOrder(ctx, nil)
	_, errs["ConvertFund"] = client.ConvertFund(ctx, nil)
	_, errs["CreateMIDVA"] = client.CreateMIDVA(ctx, nil)
	_, errs["SetWebhookURL"] = client.SetWebhookURL(ctx, nil)
	_, errs["SubmitKYB"] = client.SubmitKYB(ctx, nil)
	_, errs["CreateKYC"] = client.CreateKYC(ctx, nil)

	for name, err := range errs {
		if !errors.Is(err, goholdstation.ErrNilRequest) {
			t.Errorf("%s: error = %v, want ErrNilRequest", name, err)
		}
	}

	if api.calls != 0 {
		t.Errorf("server received %d requests, want 0", api.calls)
	}
}

func TestNewClient_BaseURLMustBeHTTPS(t *testing.T) {
	t.Parallel()

	rejected := []string{
		"http://payment.capybera.xyz",
		"http://localhost.evil.example",
		"ftp://payment.capybera.xyz",
		"payment.capybera.xyz",
		"https://user:pass@payment.capybera.xyz",
		"https://user@payment.capybera.xyz",
		"https://payment.capybera.xyz?x=1",
		"https://payment.capybera.xyz?",
		"https://payment.capybera.xyz#frag",
		"https://",
		"://bad",
	}

	for _, baseURL := range rejected {
		if _, err := goholdstation.NewClient(baseURL, testAppKey); !errors.Is(err, goholdstation.ErrInsecureBaseURL) {
			t.Errorf("%s: error = %v, want ErrInsecureBaseURL", baseURL, err)
		}
	}

	accepted := []string{
		goholdstation.SandboxBaseURL,
		goholdstation.ProductionBaseURL,
		"https://payment.capybera.xyz/",
		"https://proxy.example/holdstation",
		"http://127.0.0.1:8080",
		"http://localhost:8080",
		"http://[::1]:8080",
	}

	for _, baseURL := range accepted {
		if _, err := goholdstation.NewClient(baseURL, testAppKey); err != nil {
			t.Errorf("%s: error = %v, want none", baseURL, err)
		}
	}
}

func TestClient_DefaultHTTPClientDoesNotFollowRedirects(t *testing.T) {
	t.Parallel()

	var targetCalls int

	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		targetCalls++
		_, _ = w.Write([]byte(`{"message":"success","data":{"rate":{"buy":"1","sell":"1"}}}`))
	}))
	t.Cleanup(target.Close)

	redirector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL+"/stolen", http.StatusFound)
	}))
	t.Cleanup(redirector.Close)

	client, err := goholdstation.NewClient(redirector.URL, testAppKey)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	resp, err := client.GetExchangeRate(context.Background(), nil)
	if !errors.Is(err, goholdstation.ErrInvalidResponse) {
		t.Errorf("resp = %+v, error = %v, want ErrInvalidResponse for the 302", resp, err)
	}

	if targetCalls != 0 {
		t.Errorf("redirect target received %d requests, want 0", targetCalls)
	}
}

func TestClient_ResponseBodyIsCapped(t *testing.T) {
	t.Parallel()

	for _, status := range []int{http.StatusOK, http.StatusBadGateway} {
		client, _ := newFakeAPI(t, status, strings.Repeat("a", goholdstation.MaxResponseBytes+1))

		_, err := client.GetExchangeRate(context.Background(), nil)
		if !errors.Is(err, goholdstation.ErrResponseTooLarge) || !errors.Is(err, goholdstation.ErrInvalidResponse) {
			t.Errorf("status %d: error = %.200v, want ErrResponseTooLarge", status, err)
		}
	}
}

func TestClient_StatusOutside2xxBelow400IsInvalidResponse(t *testing.T) {
	t.Parallel()

	for _, status := range []int{http.StatusFound, http.StatusNotModified, http.StatusPermanentRedirect} {
		client, _ := newFakeAPI(t, status, `{}`)

		resp, err := client.GetExchangeRate(context.Background(), nil)
		if resp != nil || !errors.Is(err, goholdstation.ErrInvalidResponse) {
			t.Errorf("status %d: resp = %+v, error = %v, want ErrInvalidResponse", status, resp, err)
		}

		var apiErr *goholdstation.APIError
		if errors.As(err, &apiErr) {
			t.Errorf("status %d: error is an *APIError", status)
		}
	}
}
