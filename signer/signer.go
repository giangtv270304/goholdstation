// Package signer implements the Holdstation Pay Ed25519 request signature.
package signer

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	HeaderSignature = "X-HSPay-Signature"
	HeaderTimestamp = "X-HSPay-Timestamp"

	pkcs8PrivateKeySize = 48
	spkiPublicKeySize   = 44
)

var (
	// ErrInvalidPrivateKey never carries the key material.
	ErrInvalidPrivateKey = errors.New("signer: private key must be base64 of a 64-byte key, a 32-byte seed or a 48-byte PKCS#8 DER")
	// ErrInvalidPublicKey is returned for a key that is not base64 of 32 raw bytes or 44 SPKI DER bytes.
	ErrInvalidPublicKey = errors.New("signer: public key must be base64 of 32 raw bytes or a 44-byte SPKI DER")
	// ErrInvalidSignature is returned when a signature is not valid base64 or does not match.
	ErrInvalidSignature = errors.New("signer: signature does not match")
)

// Params is everything the signature covers.
// Query is the canonical query from CanonicalQuery. Body is the exact bytes sent on the wire.
type Params struct {
	Query     string
	Body      []byte
	Timestamp time.Time
}

// CanonicalQuery returns "k1=v1&k2=v2" with keys sorted, values not URL-encoded, first value of each key.
func CanonicalQuery(query url.Values) string {
	keys := make([]string, 0, len(query))
	for key := range query {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	pairs := make([]string, 0, len(keys))
	for _, key := range keys {
		pairs = append(pairs, key+"="+query.Get(key))
	}

	return strings.Join(pairs, "&")
}

// Message returns the string that gets signed: query, body and Unix-seconds timestamp joined by "\n".
func Message(params Params) string {
	return params.Query + "\n" + string(params.Body) + "\n" + strconv.FormatInt(params.Timestamp.Unix(), 10)
}

// ParsePrivateKey decodes a base64 (standard alphabet) Ed25519 private key:
// 64 bytes (seed followed by public key), a 32-byte seed, or a 48-byte PKCS#8 DER.
func ParsePrivateKey(encoded string) (ed25519.PrivateKey, error) {
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil {
		return nil, ErrInvalidPrivateKey
	}

	switch len(raw) {
	case ed25519.PrivateKeySize:
		return ed25519.PrivateKey(raw), nil
	case ed25519.SeedSize:
		return ed25519.NewKeyFromSeed(raw), nil
	case pkcs8PrivateKeySize:
		parsed, err := x509.ParsePKCS8PrivateKey(raw)
		if err != nil {
			return nil, ErrInvalidPrivateKey
		}

		key, ok := parsed.(ed25519.PrivateKey)
		if !ok {
			return nil, ErrInvalidPrivateKey
		}

		return key, nil
	default:
		return nil, ErrInvalidPrivateKey
	}
}

// ParsePublicKey decodes a base64 (standard alphabet) Ed25519 public key: 32 raw bytes or a 44-byte SPKI DER.
func ParsePublicKey(encoded string) (ed25519.PublicKey, error) {
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil {
		return nil, ErrInvalidPublicKey
	}

	switch len(raw) {
	case ed25519.PublicKeySize:
		return ed25519.PublicKey(raw), nil
	case spkiPublicKeySize:
		parsed, err := x509.ParsePKIXPublicKey(raw)
		if err != nil {
			return nil, ErrInvalidPublicKey
		}

		key, ok := parsed.(ed25519.PublicKey)
		if !ok {
			return nil, ErrInvalidPublicKey
		}

		return key, nil
	default:
		return nil, ErrInvalidPublicKey
	}
}

// Sign returns the standard base64 Ed25519 signature of message.
func Sign(privateKey ed25519.PrivateKey, message string) string {
	return base64.StdEncoding.EncodeToString(ed25519.Sign(privateKey, []byte(message)))
}

// Verify checks a standard base64 Ed25519 signature over message.
func Verify(publicKey ed25519.PublicKey, message []byte, signature string) error {
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(signature))
	if err != nil {
		return fmt.Errorf("%w: not valid base64", ErrInvalidSignature)
	}

	if !ed25519.Verify(publicKey, message, raw) {
		return ErrInvalidSignature
	}

	return nil
}

// Headers returns the two signature headers to add to a signed request.
func Headers(privateKey ed25519.PrivateKey, params Params) map[string]string {
	return map[string]string{
		HeaderSignature: Sign(privateKey, Message(params)),
		HeaderTimestamp: strconv.FormatInt(params.Timestamp.Unix(), 10),
	}
}
