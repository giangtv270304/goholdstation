package goholdstation

import (
	"context"
	"net/http"
)

// DecodeQRRequest: set exactly one of QRString and QRImageURL (HTTPS).
type DecodeQRRequest struct {
	QRString   string `json:"qr_string,omitempty"`
	QRImageURL string `json:"qr_image_url,omitempty"`
}

// QRData is the content of a decoded bank QR code.
type QRData struct {
	BankBinCode       string `json:"bank_bin_code"`
	BankAccountNumber string `json:"bank_account_number"`
	// Amount is in VND, "0" when the QR code carries no amount.
	Amount  Decimal `json:"amount"`
	Purpose string  `json:"purpose"`
}

// DecodeQRResponse: a failed decode is HTTP 200 with Data nil and ErrorCode 1 to 4; it is not an *APIError.
type DecodeQRResponse struct {
	Data  *QRData `json:"data"`
	Error string  `json:"error"`
	// ErrorCode: 1 invalid input, 2 image download failed, 3 image decode failed, 4 parse failed.
	ErrorCode int    `json:"error_code"`
	Message   string `json:"message"`
}

// DecodeQR decodes a bank QR code. Signed. Check Data and ErrorCode in the result.
func (c *Client) DecodeQR(ctx context.Context, req *DecodeQRRequest) (*Response[DecodeQRResponse], error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	return call[DecodeQRResponse](ctx, c, request{
		Method:               http.MethodPost,
		Path:                 "/partners/payment-methods/qr/decode",
		Query:                nil,
		Body:                 req,
		Signed:               true,
		AllowBusinessFailure: true,
		DataKeyInPayload:     true,
	})
}
