// Package goholdstation is a Go client for the Holdstation Pay partner API.
//
//	client, err := goholdstation.NewClient(goholdstation.SandboxBaseURL, partnerAppKey,
//		goholdstation.WithPrivateKey(ed25519PrivateKey))
//	if err != nil {
//		return err
//	}
//
//	resp, err := client.GetExchangeRate(ctx, nil)
package goholdstation

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/giangtv270304/goholdstation/signer"
)

const (
	// Version is the SDK version, sent in the User-Agent header.
	Version = "0.1.1"

	SandboxBaseURL    = "https://payment.capybera.xyz"
	ProductionBaseURL = "https://pay-api.holdstation.com"

	// DefaultTimeout applies to the default HTTP client.
	DefaultTimeout = 30 * time.Second

	// HeaderPartnerAppKey carries the partner API key on every request.
	HeaderPartnerAppKey = "Partner-App-Key"

	// MaxResponseBytes is the largest response body the client reads; a larger one is an error.
	MaxResponseBytes = 10 << 20

	userAgent = "goholdstation/" + Version
)

var (
	ErrEmptyBaseURL = errors.New("goholdstation: baseURL is required")
	ErrEmptyAppKey  = errors.New("goholdstation: partnerAppKey is required")
	// ErrInsecureBaseURL is returned for a base URL that is not https (http is allowed on a loopback host only),
	// or that carries userinfo, a query or a fragment.
	ErrInsecureBaseURL = errors.New("goholdstation: baseURL must be an https URL without userinfo, query or fragment")
	ErrNilRequest      = errors.New("goholdstation: request is nil")
	// ErrInvalidResponse is wrapped by every error about a response the client cannot use: a status below 400
	// that is not 2xx (redirects are not followed), a 2xx body that does not decode, or a body that is too large.
	ErrInvalidResponse = errors.New("goholdstation: invalid response")
	// ErrResponseTooLarge is returned, together with ErrInvalidResponse, when the body is larger than MaxResponseBytes.
	ErrResponseTooLarge = errors.New("goholdstation: response body is larger than MaxResponseBytes")
	// ErrMissingField is wrapped with the name of the empty required field.
	ErrMissingField = errors.New("goholdstation: required field is empty")
	// ErrMissingPrivateKey is returned, before any network call, by a signed operation on a client without WithPrivateKey.
	ErrMissingPrivateKey = errors.New("goholdstation: this operation is signed: create the client with WithPrivateKey")
)

// HTTPClient is satisfied by *http.Client and by a test double.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Logger is satisfied by *slog.Logger.
type Logger interface {
	Debug(msg string, args ...any)
}

// Option configures a Client in NewClient.
type Option func(*Client)

// Client calls the Holdstation Pay partner API. Create one with NewClient and reuse it.
// It never retries a request.
type Client struct {
	baseURL       string
	partnerAppKey string
	privateKeyB64 string
	privateKey    ed25519.PrivateKey
	httpClient    HTTPClient
	logger        Logger
	now           func() time.Time
}

// WithPrivateKey sets the Ed25519 request-signing key, required by the signed operations.
// It is base64 (standard alphabet) of the 64-byte key, the 32-byte seed or the 48-byte PKCS#8 DER.
// The key is for servers only: never ship it in a browser or a mobile app.
func WithPrivateKey(privateKey string) Option {
	return func(c *Client) { c.privateKeyB64 = privateKey }
}

// WithHTTPClient replaces the default client (timeout DefaultTimeout, redirects not followed).
// A custom client should not follow redirects either: a redirect would carry the app key to another host.
func WithHTTPClient(httpClient HTTPClient) Option {
	return func(c *Client) { c.httpClient = httpClient }
}

// WithLogger turns on debug logs of method, path and status. Query strings, bodies, headers and keys are never logged.
func WithLogger(logger Logger) Option {
	return func(c *Client) { c.logger = logger }
}

// WithClock replaces the clock used for the signature timestamp.
// The server rejects a timestamp more than 60 seconds away from its own time.
func WithClock(now func() time.Time) Option {
	return func(c *Client) { c.now = now }
}

// NewClient returns a client for baseURL (SandboxBaseURL or ProductionBaseURL).
// It fails when baseURL or partnerAppKey is empty, when baseURL is not https (ErrInsecureBaseURL),
// or when the key given to WithPrivateKey cannot be decoded.
// The partner app key is for servers only.
func NewClient(baseURL, partnerAppKey string, opts ...Option) (*Client, error) {
	if baseURL == "" {
		return nil, ErrEmptyBaseURL
	}

	if partnerAppKey == "" {
		return nil, ErrEmptyAppKey
	}

	if err := checkBaseURL(baseURL); err != nil {
		return nil, err
	}

	client := &Client{
		baseURL:       strings.TrimRight(baseURL, "/"),
		partnerAppKey: partnerAppKey,
		privateKeyB64: "",
		privateKey:    nil,
		httpClient:    newDefaultHTTPClient(),
		logger:        nil,
		now:           time.Now,
	}

	for _, opt := range opts {
		opt(client)
	}

	if client.privateKeyB64 != "" {
		privateKey, err := signer.ParsePrivateKey(client.privateKeyB64)
		if err != nil {
			return nil, fmt.Errorf("goholdstation: WithPrivateKey: %w", err)
		}

		client.privateKey = privateKey
		client.privateKeyB64 = ""
	}

	return client, nil
}

func checkBaseURL(baseURL string) error {
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" {
		return ErrInsecureBaseURL
	}

	host := parsed.Hostname()
	loopback := host == "localhost" || host == "127.0.0.1" || host == "::1"

	if parsed.Scheme != "https" && (parsed.Scheme != "http" || !loopback) {
		return ErrInsecureBaseURL
	}

	return nil
}

func newDefaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout:       DefaultTimeout,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
}

// String hides the keys, so that printing a Client never leaks them.
func (c *Client) String() string {
	return "goholdstation.Client{baseURL: " + c.baseURL + "}"
}

// GoString hides the keys for the %#v verb.
func (c *Client) GoString() string {
	return c.String()
}

func call[T any](ctx context.Context, c *Client, req request) (*Response[T], error) {
	resp := new(Response[T])

	message, err := c.execute(ctx, req, &resp.Data)
	if err != nil {
		return nil, err
	}

	resp.Message = message

	return resp, nil
}

func (c *Client) execute(ctx context.Context, req request, out any) (string, error) {
	if req.Signed && c.privateKey == nil {
		return "", fmt.Errorf("%w (%s %s)", ErrMissingPrivateKey, req.Method, req.Path)
	}

	var body []byte

	if req.Body != nil {
		encoded, err := json.Marshal(req.Body)
		if err != nil {
			return "", fmt.Errorf("goholdstation: encode request body: %w", err)
		}

		body = encoded
	}

	query := req.Query.Encode()

	fullURL := c.baseURL + req.Path
	if query != "" {
		fullURL += "?" + query
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, fullURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("goholdstation: build request: %w", err)
	}

	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("User-Agent", userAgent)
	httpReq.Header.Set(HeaderPartnerAppKey, c.partnerAppKey)

	if body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	if req.Signed {
		// The signature covers the exact bytes sent: sign after encoding, never re-encode.
		signed := signer.Headers(c.privateKey, signer.Params{
			Query:     signer.CanonicalQuery(req.Query),
			Body:      body,
			Timestamp: c.now(),
		})
		for key, value := range signed {
			httpReq.Header.Set(key, value)
		}
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("goholdstation: send %s %s: %w", req.Method, req.Path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, MaxResponseBytes+1))
	if err != nil {
		return "", fmt.Errorf("goholdstation: read response of %s %s: %w", req.Method, req.Path, err)
	}

	if len(raw) > MaxResponseBytes {
		return "", fmt.Errorf("%w: %w (%s %s, status %d)", ErrInvalidResponse, ErrResponseTooLarge, req.Method, req.Path, resp.StatusCode)
	}

	if c.logger != nil {
		c.logger.Debug("goholdstation request", "method", req.Method, "path", req.Path, "status", resp.StatusCode)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return "", newAPIError(resp.StatusCode, raw)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("%w: unexpected status %d of %s %s", ErrInvalidResponse, resp.StatusCode, req.Method, req.Path)
	}

	message, payload, err := splitEnvelope(raw, req.DataKeyInPayload)
	if err != nil {
		return "", fmt.Errorf("%w: decode response of %s %s (status %d): %w", ErrInvalidResponse, req.Method, req.Path, resp.StatusCode, err)
	}

	if !req.AllowBusinessFailure {
		if apiErr := businessFailure(resp.StatusCode, raw, payload); apiErr != nil {
			return "", apiErr
		}
	}

	if err := json.Unmarshal(payload, out); err != nil {
		return "", fmt.Errorf("%w: decode response of %s %s (status %d): %w", ErrInvalidResponse, req.Method, req.Path, resp.StatusCode, err)
	}

	return message, nil
}
