package goholdstation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// Values of KYCInput.IDType.
const (
	KYCIDTypeCitizenID     = 1
	KYCIDTypePassport      = 2
	KYCIDTypeDriverLicense = 3
)

// Values of KYCInput.Sex.
const (
	KYCSexMale   = 1
	KYCSexFemale = 2
	KYCSexOther  = 3
)

// KYCImageInput is one uploaded image. At most 10 MB once decoded.
type KYCImageInput struct {
	// ContentType is "image/jpeg", "image/png" or "image/heic".
	ContentType string `json:"content_type"`
	// Data is standard base64 without a data-URI prefix.
	Data string `json:"data"`
}

// KYCInput is the body of CreateKYC and UpdateKYC. Holdstation configures the required fields per partner.
// Empty fields are left out. The whole body is limited to 32 MB.
type KYCInput struct {
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Address  string `json:"address,omitempty"`
	IDNumber string `json:"id_number,omitempty"`
	// IDType is 1 Citizen ID, 2 Passport or 3 Driver License.
	IDType int `json:"id_type,omitempty"`
	// Sex is 1 Male, 2 Female or 3 Other.
	Sex         int    `json:"sex,omitempty"`
	Nationality string `json:"nationality,omitempty"`
	// DateOfIssue and DateOfExpiry are "YYYY-MM-DD". Leave DateOfExpiry empty when the document does not expire.
	DateOfIssue  string `json:"date_of_issue,omitempty"`
	DateOfExpiry string `json:"date_of_expiry,omitempty"`
	// On update, a nil image keeps the stored one; images cannot be removed.
	IDFrontImg *KYCImageInput `json:"id_front_img,omitempty"`
	IDBackImg  *KYCImageInput `json:"id_back_img,omitempty"`
	SelfieImg  *KYCImageInput `json:"selfie_img,omitempty"`
}

// KYCImage is a stored image. The API never returns its content.
type KYCImage struct {
	FileRef     string `json:"file_ref"`
	ContentType string `json:"content_type"`
	// Size is in bytes.
	Size int64 `json:"size"`
}

// KYC is a customer identity record.
type KYC struct {
	// ID is the id chosen by the partner; use it as kyc_id on orders.
	ID           string     `json:"id"`
	PartnerID    string     `json:"partner_id"`
	Name         string     `json:"name"`
	Email        string     `json:"email"`
	Phone        string     `json:"phone"`
	Address      string     `json:"address"`
	IDNumber     string     `json:"id_number"`
	IDType       int        `json:"id_type"`
	Sex          int        `json:"sex"`
	Nationality  string     `json:"nationality"`
	DateOfIssue  string     `json:"date_of_issue"`
	DateOfExpiry string     `json:"date_of_expiry"`
	CreatedAt    *Timestamp `json:"created_at"`
	UpdatedAt    *Timestamp `json:"updated_at"`
	IDFrontImg   *KYCImage  `json:"id_front_img"`
	IDBackImg    *KYCImage  `json:"id_back_img"`
	SelfieImg    *KYCImage  `json:"selfie_img"`
}

// KYCResult is the response of CreateKYC, UpdateKYC and GetKYCByID.
// The API documents two forms, {"kyc": {...}} and the record itself; both fill KYC.
type KYCResult struct {
	KYC *KYC `json:"kyc"`
}

// UnmarshalJSON accepts {"success": true, "kyc": {...}} and the flat record.
func (r *KYCResult) UnmarshalJSON(data []byte) error {
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(data, &keys); err != nil {
		return err
	}

	record := json.RawMessage(data)
	if nested, ok := keys["kyc"]; ok {
		record = nested
	} else if _, flat := keys["id"]; !flat {
		r.KYC = nil

		return nil
	}

	var kyc *KYC
	if err := json.Unmarshal(record, &kyc); err != nil {
		return err
	}

	r.KYC = kyc

	return nil
}

type GetKYCByIDRequest struct {
	ID string
}

type GetKYCByIDResponse = KYCResult

// GetKYCByID returns one KYC record. Signed. An unknown id is a 404 *APIError with code "KYC_NOT_FOUND".
// It fails with ErrMissingField when ID is empty.
func (c *Client) GetKYCByID(ctx context.Context, req *GetKYCByIDRequest) (*Response[GetKYCByIDResponse], error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	if req.ID == "" {
		return nil, fmt.Errorf("%w: ID", ErrMissingField)
	}

	return call[GetKYCByIDResponse](ctx, c, request{
		Method: http.MethodGet,
		Path:   "/partners/kycs/" + url.PathEscape(req.ID),
		Query:  nil,
		Body:   nil,
		Signed: true,
	})
}
