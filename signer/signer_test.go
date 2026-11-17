package signer_test

import (
	"bytes"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/giangtv270304/goholdstation/signer"
)

type vectorFile struct {
	PrivateKey string `json:"privateKey"`
	Seed       string `json:"seed"`
	PublicKey  string `json:"publicKey"`
	Primitive  []struct {
		Name      string `json:"name"`
		Source    string `json:"source"`
		Message   string `json:"message"`
		Signature string `json:"signature"`
	} `json:"primitive"`
	Requests []struct {
		Name      string `json:"name"`
		Source    string `json:"source"`
		Query     string `json:"query"`
		Body      string `json:"body"`
		Timestamp int64  `json:"timestamp"`
		Signature string `json:"signature"`
	} `json:"requests"`
}

func loadVectors(t *testing.T) vectorFile {
	t.Helper()

	raw, err := os.ReadFile("signature-vectors.json")
	if err != nil {
		t.Fatalf("read vectors: %v", err)
	}

	var vectors vectorFile
	if err := json.Unmarshal(raw, &vectors); err != nil {
		t.Fatalf("decode vectors: %v", err)
	}

	if len(vectors.Primitive) == 0 || len(vectors.Requests) == 0 {
		t.Fatal("vector file has no vectors")
	}

	return vectors
}

func TestSignVectors(t *testing.T) {
	t.Parallel()

	vectors := loadVectors(t)

	privateKey, err := signer.ParsePrivateKey(vectors.PrivateKey)
	if err != nil {
		t.Fatalf("ParsePrivateKey: %v", err)
	}

	publicKey, err := signer.ParsePublicKey(vectors.PublicKey)
	if err != nil {
		t.Fatalf("ParsePublicKey: %v", err)
	}

	for _, testCase := range vectors.Primitive {
		t.Run(testCase.Source+"/"+testCase.Name, func(t *testing.T) {
			t.Parallel()

			if got := signer.Sign(privateKey, testCase.Message); got != testCase.Signature {
				t.Errorf("signature = %q, want %q", got, testCase.Signature)
			}

			if err := signer.Verify(publicKey, []byte(testCase.Message), testCase.Signature); err != nil {
				t.Errorf("Verify: %v", err)
			}
		})
	}

	for _, testCase := range vectors.Requests {
		t.Run(testCase.Source+"/"+testCase.Name, func(t *testing.T) {
			t.Parallel()

			headers := signer.Headers(privateKey, signer.Params{
				Query:     testCase.Query,
				Body:      []byte(testCase.Body),
				Timestamp: time.Unix(testCase.Timestamp, 0),
			})

			if got := headers[signer.HeaderSignature]; got != testCase.Signature {
				t.Errorf("signature = %q, want %q", got, testCase.Signature)
			}

			if got := headers[signer.HeaderTimestamp]; got != "1700000000" {
				t.Errorf("timestamp = %q, want 1700000000", got)
			}
		})
	}
}

func TestParsePrivateKey_AcceptedEncodings(t *testing.T) {
	t.Parallel()

	vectors := loadVectors(t)

	want, err := signer.ParsePrivateKey(vectors.PrivateKey)
	if err != nil {
		t.Fatalf("ParsePrivateKey: %v", err)
	}

	der, err := x509.MarshalPKCS8PrivateKey(want)
	if err != nil {
		t.Fatalf("MarshalPKCS8PrivateKey: %v", err)
	}

	testCases := map[string]string{
		"32-byte seed":      vectors.Seed,
		"48-byte PKCS#8":    base64.StdEncoding.EncodeToString(der),
		"64-byte with trim": " " + vectors.PrivateKey + "\n",
	}

	for name, encoded := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := signer.ParsePrivateKey(encoded)
			if err != nil {
				t.Fatalf("ParsePrivateKey: %v", err)
			}

			if !got.Equal(want) {
				t.Error("decoded key differs from the 64-byte form")
			}
		})
	}

	derived, ok := want.Public().(ed25519.PublicKey)
	if !ok {
		t.Fatal("public key has an unexpected type")
	}

	if got := base64.StdEncoding.EncodeToString(derived); got != vectors.PublicKey {
		t.Errorf("public key = %q, want %q", got, vectors.PublicKey)
	}
}

func TestParsePrivateKey_InvalidNeverLeaksInput(t *testing.T) {
	t.Parallel()

	secret := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{7}, 31))

	for _, input := range []string{"", "not base64 !!", secret, base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{7}, 48))} {
		_, err := signer.ParsePrivateKey(input)
		if !errors.Is(err, signer.ErrInvalidPrivateKey) {
			t.Fatalf("error = %v, want ErrInvalidPrivateKey", err)
		}

		if input != "" && strings.Contains(err.Error(), input) {
			t.Errorf("error message contains the key input: %v", err)
		}
	}
}

func TestParsePublicKey_RawAndSPKI(t *testing.T) {
	t.Parallel()

	publicKey, _, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	der, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		t.Fatalf("MarshalPKIXPublicKey: %v", err)
	}

	for name, raw := range map[string][]byte{"raw": publicKey, "spki": der} {
		got, err := signer.ParsePublicKey(base64.StdEncoding.EncodeToString(raw))
		if err != nil {
			t.Fatalf("%s: ParsePublicKey: %v", name, err)
		}

		if !got.Equal(publicKey) {
			t.Errorf("%s: decoded key differs", name)
		}
	}

	if _, err := signer.ParsePublicKey("AAAA"); !errors.Is(err, signer.ErrInvalidPublicKey) {
		t.Errorf("error = %v, want ErrInvalidPublicKey", err)
	}
}

func TestCanonicalQuery_SortedAndUnescaped(t *testing.T) {
	t.Parallel()

	query := url.Values{}
	query.Set("token", "USD T")
	query.Set("page", "1")
	query.Set("change_type", "1")

	if got, want := signer.CanonicalQuery(query), "change_type=1&page=1&token=USD T"; got != want {
		t.Errorf("canonical query = %q, want %q", got, want)
	}

	if got := signer.CanonicalQuery(nil); got != "" {
		t.Errorf("canonical query of nil = %q, want empty", got)
	}
}

func TestVerify_RejectsChangedMessageAndBadBase64(t *testing.T) {
	t.Parallel()

	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	signature := signer.Sign(privateKey, "message")

	if err := signer.Verify(publicKey, []byte("message"), signature); err != nil {
		t.Fatalf("Verify: %v", err)
	}

	if err := signer.Verify(publicKey, []byte("messagf"), signature); !errors.Is(err, signer.ErrInvalidSignature) {
		t.Errorf("error = %v, want ErrInvalidSignature", err)
	}

	if err := signer.Verify(publicKey, []byte("message"), "%%%"); !errors.Is(err, signer.ErrInvalidSignature) {
		t.Errorf("error = %v, want ErrInvalidSignature", err)
	}
}
