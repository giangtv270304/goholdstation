package goholdstation_test

import (
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

const (
	envBaseURL          = "HOLDSTATION_BASE_URL"
	envPartnerAppKey    = "HOLDSTATION_PARTNER_APP_KEY"
	envPrivateKey       = "HOLDSTATION_ED25519_PRIVATE_KEY"
	envTestChainID      = "HOLDSTATION_TEST_CHAIN_ID"
	envTestTokenAddress = "HOLDSTATION_TEST_TOKEN_ADDRESS"
	envTestRecipient    = "HOLDSTATION_TEST_RECIPIENT"
	envTestCallbackURL  = "HOLDSTATION_TEST_CALLBACK_URL"
	envTestKYCID        = "HOLDSTATION_TEST_KYC_ID"
	envTestBankBinCode  = "HOLDSTATION_TEST_BANK_BIN_CODE"
	envTestBankAccount  = "HOLDSTATION_TEST_BANK_ACCOUNT_NUMBER"

	testAppKey = "test-partner-app-key"

	unauthorizedBody = `{"error":"unauthorized","error_code":"UNAUTHORIZED","message":"Invalid or missing API key"}`
	notFoundBody     = `{"error":"not found","error_code":"NOT_FOUND","message":"No fund log with this ID"}`
	badRequestBody   = `{"error":"bad request","error_code":"BAD_REQUEST","message":"invalid request"}`
)

type fakeAPI struct {
	publicKey ed25519.PublicKey
	status    int
	body      string

	mu           sync.Mutex
	calls        int
	method       string
	path         string
	rawQuery     string
	requestBody  string
	header       http.Header
	signatureErr string
}

type wantRequest struct {
	method string
	path   string
	query  string
	body   string
	signed bool
}

func newTestKey(t *testing.T) (ed25519.PublicKey, string) {
	t.Helper()

	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	return publicKey, base64.StdEncoding.EncodeToString(privateKey)
}

func newFakeAPI(t *testing.T, status int, body string, opts ...goholdstation.Option) (*goholdstation.Client, *fakeAPI) {
	t.Helper()

	publicKey, privateKey := newTestKey(t)
	api := &fakeAPI{publicKey: publicKey, status: status, body: body}

	server := httptest.NewServer(api)
	t.Cleanup(server.Close)

	options := append([]goholdstation.Option{goholdstation.WithPrivateKey(privateKey)}, opts...)

	client, err := goholdstation.NewClient(server.URL, testAppKey, options...)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	return client, api
}

func (f *fakeAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	raw, _ := io.ReadAll(r.Body)

	f.mu.Lock()
	f.calls++
	f.method = r.Method
	f.path = r.URL.EscapedPath()
	f.rawQuery = r.URL.RawQuery
	f.requestBody = string(raw)
	f.header = r.Header.Clone()
	f.signatureErr = f.verifySignature(r, raw)
	f.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(f.status)
	_, _ = io.WriteString(w, f.body)
}

func (f *fakeAPI) verifySignature(r *http.Request, body []byte) string {
	signature := r.Header.Get("X-HSPay-Signature")
	timestamp := r.Header.Get("X-HSPay-Timestamp")

	if signature == "" && timestamp == "" {
		return "signature headers are absent"
	}

	if _, err := strconv.ParseInt(timestamp, 10, 64); err != nil {
		return "X-HSPay-Timestamp is not Unix seconds: " + timestamp
	}

	query := r.URL.Query()

	keys := make([]string, 0, len(query))
	for key := range query {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	pairs := make([]string, 0, len(keys))
	for _, key := range keys {
		pairs = append(pairs, key+"="+query.Get(key))
	}

	message := strings.Join(pairs, "&") + "\n" + string(body) + "\n" + timestamp

	decoded, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return "X-HSPay-Signature is not standard base64"
	}

	if !ed25519.Verify(f.publicKey, []byte(message), decoded) {
		return "Ed25519 signature does not match the request received"
	}

	return ""
}

func (f *fakeAPI) assertRequest(t *testing.T, want wantRequest) {
	t.Helper()

	f.mu.Lock()
	defer f.mu.Unlock()

	if f.calls != 1 {
		t.Fatalf("server received %d requests, want 1", f.calls)
	}

	if f.method != want.method || f.path != want.path {
		t.Errorf("request = %s %s, want %s %s", f.method, f.path, want.method, want.path)
	}

	if f.rawQuery != want.query {
		t.Errorf("query = %q, want %q", f.rawQuery, want.query)
	}

	if f.requestBody != want.body {
		t.Errorf("body = %s, want %s", f.requestBody, want.body)
	}

	if got := f.header.Get("Partner-App-Key"); got != testAppKey {
		t.Errorf("Partner-App-Key = %q, want %q", got, testAppKey)
	}

	if got, wantUA := f.header.Get("User-Agent"), "goholdstation/"+goholdstation.Version; got != wantUA {
		t.Errorf("User-Agent = %q, want %q", got, wantUA)
	}

	if got := f.header.Get("Content-Type"); want.body != "" && got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}

	hasSignatureHeaders := f.header.Get("X-HSPay-Signature") != "" || f.header.Get("X-HSPay-Timestamp") != ""

	if want.signed && f.signatureErr != "" {
		t.Errorf("signed request: %s", f.signatureErr)
	}

	if !want.signed && hasSignatureHeaders {
		t.Error("unsigned operation carries signature headers")
	}
}

func assertAPIError(t *testing.T, err error, status int, code string) {
	t.Helper()

	_ = requireAPIError(t, err, status, code)
}

func requireAPIError(t *testing.T, err error, status int, code string) *goholdstation.APIError {
	t.Helper()

	var apiErr *goholdstation.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %v, want *goholdstation.APIError", err)
	}

	if apiErr.StatusCode != status || apiErr.Code != code {
		t.Errorf("api error = status %d code %q, want status %d code %q", apiErr.StatusCode, apiErr.Code, status, code)
	}

	if len(apiErr.Body) == 0 {
		t.Error("api error has no raw body")
	}

	return apiErr
}

func assertMissingField(t *testing.T, err error, api *fakeAPI) {
	t.Helper()

	if !errors.Is(err, goholdstation.ErrMissingField) {
		t.Errorf("error = %v, want ErrMissingField", err)
	}

	api.mu.Lock()
	defer api.mu.Unlock()

	if api.calls != 0 {
		t.Errorf("server received %d requests, want 0", api.calls)
	}
}

func decimalPtr(value string) *goholdstation.Decimal {
	decimal := goholdstation.Decimal(value)

	return &decimal
}

func sandboxBaseURL(t *testing.T) string {
	t.Helper()

	baseURL := os.Getenv(envBaseURL)
	if baseURL == "" {
		baseURL = goholdstation.SandboxBaseURL
	}

	if strings.TrimRight(strings.ToLower(baseURL), "/") == goholdstation.ProductionBaseURL {
		t.Fatalf("%s points at production; sandbox tests never run there", envBaseURL)
	}

	return baseURL
}

func newSandboxClient(t *testing.T) *goholdstation.Client {
	t.Helper()

	appKey := os.Getenv(envPartnerAppKey)
	if appKey == "" {
		t.Skipf("sandbox test skipped: set %s", envPartnerAppKey)
	}

	client, err := goholdstation.NewClient(sandboxBaseURL(t), appKey)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	return client
}

func newSignedSandboxClient(t *testing.T) *goholdstation.Client {
	t.Helper()

	appKey := os.Getenv(envPartnerAppKey)
	privateKey := os.Getenv(envPrivateKey)

	if appKey == "" || privateKey == "" {
		t.Skipf("sandbox test skipped: set %s and %s", envPartnerAppKey, envPrivateKey)
	}

	client, err := goholdstation.NewClient(sandboxBaseURL(t), appKey, goholdstation.WithPrivateKey(privateKey))
	if err != nil {
		t.Fatal("NewClient: the sandbox private key cannot be decoded")
	}

	return client
}

func requireEnv(t *testing.T, names ...string) map[string]string {
	t.Helper()

	values := make(map[string]string, len(names))

	for _, name := range names {
		value := os.Getenv(name)
		if value == "" {
			t.Skipf("sandbox test skipped: set %s", name)
		}

		values[name] = value
	}

	return values
}

func envOrEmpty(name string) string {
	return os.Getenv(name)
}
