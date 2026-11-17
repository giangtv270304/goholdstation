package goholdstation

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// KYBStatus is the review state of a KYB record. Unknown values decode without error.
type KYBStatus int32

const (
	KYBStatusPending  KYBStatus = 1
	KYBStatusApproved KYBStatus = 2
	KYBStatusRejected KYBStatus = 3
)

// Values of KYBSubmission.SubjectType.
const (
	KYBSubjectTypePartner     = 1
	KYBSubjectTypeSubMerchant = 2
)

// Values of KYBPersonInput.Role.
const (
	KYBPersonRoleLegalRepresentative = 1
	KYBPersonRoleDirector            = 2
)

// KYBDocumentType is the kind of a KYB document.
type KYBDocumentType int32

const (
	KYBDocumentTypeIncorporationPaper          KYBDocumentType = 1
	KYBDocumentTypeBusinessRegistrationExtract KYBDocumentType = 2
	// KYBDocumentTypeBusinessDeclaration: stamped declaration of nature of business and geographic scope.
	KYBDocumentTypeBusinessDeclaration KYBDocumentType = 3
	KYBDocumentTypeImportExportLicence KYBDocumentType = 4
	KYBDocumentTypeProductImages       KYBDocumentType = 5
	KYBDocumentTypeOther               KYBDocumentType = 9
	KYBDocumentTypePassport            KYBDocumentType = 10
	KYBDocumentTypeCitizenID           KYBDocumentType = 11
)

// KYBFileInput is one uploaded file. At most 10 MB once decoded.
type KYBFileInput struct {
	Filename string `json:"filename"`
	// ContentType is "application/pdf", "image/jpeg", "image/png" or "image/heic".
	ContentType string `json:"content_type"`
	// Data is standard base64 without a data-URI prefix. Encrypted PDFs are refused.
	Data string `json:"data"`
}

// KYBDocumentInput is one document with its files (at most 10).
type KYBDocumentInput struct {
	Type  KYBDocumentType `json:"type"`
	Files []KYBFileInput  `json:"files"`
	// DocumentNumber is required for KYBDocumentTypeImportExportLicence.
	DocumentNumber string `json:"document_number,omitempty"`
	// DateOfIssue and DateOfExpiry are "YYYY-MM-DD".
	DateOfIssue  string `json:"date_of_issue,omitempty"`
	DateOfExpiry string `json:"date_of_expiry,omitempty"`
}

// KYBPersonInput is a legal representative or director. Dates are "YYYY-MM-DD".
type KYBPersonInput struct {
	Role        int    `json:"role"`
	Name        string `json:"name"`
	DateOfBirth string `json:"date_of_birth"`
	Address     string `json:"address"`
	IDNumber    string `json:"id_number"`
	// IDType is 1 Citizen ID or 2 Passport.
	IDType int `json:"id_type"`
	// Sex is 1 Male, 2 Female or 3 Other.
	Sex          int    `json:"sex"`
	Nationality  string `json:"nationality"`
	DateOfIssue  string `json:"date_of_issue,omitempty"`
	DateOfExpiry string `json:"date_of_expiry,omitempty"`
	Email        string `json:"email,omitempty"`
	Phone        string `json:"phone,omitempty"`
	// Documents holds exactly one document of type Passport or CitizenID, matching IDType.
	Documents   []KYBDocumentInput `json:"documents"`
	IDCardFront *KYBFileInput      `json:"id_card_front,omitempty"`
	IDCardBack  *KYBFileInput      `json:"id_card_back,omitempty"`
	Selfie      *KYBFileInput      `json:"selfie,omitempty"`
}

// KYBSubmission is the body of SubmitKYB. The whole body is limited to 40 MB.
type KYBSubmission struct {
	SubjectType int `json:"subject_type"`
	// ReferenceID identifies a sub-merchant (at most 128 bytes); the API generates one when it is empty.
	ReferenceID        string `json:"reference_id,omitempty"`
	LegalName          string `json:"legal_name"`
	RegistrationNumber string `json:"registration_number"`
	Address            string `json:"address"`
	NatureOfBusiness   string `json:"nature_of_business"`
	// IncorporationCountry is ISO 3166-1 alpha-2 preferably.
	IncorporationCountry string `json:"incorporation_country"`
	// IncorporationDate is "YYYY-MM-DD".
	IncorporationDate string   `json:"incorporation_date"`
	Email             string   `json:"email,omitempty"`
	Phone             string   `json:"phone,omitempty"`
	Website           string   `json:"website,omitempty"`
	GeographicScope   []string `json:"geographic_scope"`
	// Persons: at most 20, exactly one legal representative.
	Persons []KYBPersonInput `json:"persons"`
	// Documents: at most 30; types 1 and 3 are mandatory.
	Documents []KYBDocumentInput `json:"documents"`
}

// KYBFile is a stored file. The API never returns its content.
type KYBFile struct {
	FileRef     string `json:"file_ref"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	// Size is in bytes.
	Size int64 `json:"size"`
}

// KYBDocument is a stored document.
type KYBDocument struct {
	Type           KYBDocumentType `json:"type"`
	Files          []KYBFile       `json:"files"`
	DocumentNumber string          `json:"document_number"`
	DateOfIssue    string          `json:"date_of_issue"`
	DateOfExpiry   string          `json:"date_of_expiry"`
}

// KYBPerson is a stored person.
type KYBPerson struct {
	Role         int           `json:"role"`
	Name         string        `json:"name"`
	DateOfBirth  string        `json:"date_of_birth"`
	Address      string        `json:"address"`
	IDNumber     string        `json:"id_number"`
	IDType       int           `json:"id_type"`
	Sex          int           `json:"sex"`
	Nationality  string        `json:"nationality"`
	DateOfIssue  string        `json:"date_of_issue"`
	DateOfExpiry string        `json:"date_of_expiry"`
	Email        string        `json:"email"`
	Phone        string        `json:"phone"`
	Documents    []KYBDocument `json:"documents"`
	IDCardFront  *KYBFile      `json:"id_card_front"`
	IDCardBack   *KYBFile      `json:"id_card_back"`
	Selfie       *KYBFile      `json:"selfie"`
}

// KYB is a business verification record.
type KYB struct {
	ID                   string        `json:"id"`
	PartnerID            string        `json:"partner_id"`
	SubjectType          int           `json:"subject_type"`
	ReferenceID          string        `json:"reference_id"`
	LegalName            string        `json:"legal_name"`
	RegistrationNumber   string        `json:"registration_number"`
	Address              string        `json:"address"`
	NatureOfBusiness     string        `json:"nature_of_business"`
	IncorporationCountry string        `json:"incorporation_country"`
	IncorporationDate    string        `json:"incorporation_date"`
	Email                string        `json:"email"`
	Phone                string        `json:"phone"`
	Website              string        `json:"website"`
	GeographicScope      []string      `json:"geographic_scope"`
	Persons              []KYBPerson   `json:"persons"`
	Documents            []KYBDocument `json:"documents"`
	Status               KYBStatus     `json:"status"`
	// RejectionReason is empty unless Status is KYBStatusRejected.
	RejectionReason string     `json:"rejection_reason"`
	SubmittedAt     *Timestamp `json:"submitted_at"`
	CreatedAt       *Timestamp `json:"created_at"`
	UpdatedAt       *Timestamp `json:"updated_at"`
	// ReviewedAt is nil while the record is pending.
	ReviewedAt *Timestamp `json:"reviewed_at"`
}

type GetKYBByIDRequest struct {
	// ID is the server-generated KYB id.
	ID string
}

type GetKYBByIDResponse struct {
	KYB *KYB `json:"kyb"`
}

// GetKYBByID returns one KYB record. Signed. An unknown id is a 404 *APIError.
// It fails with ErrMissingField when ID is empty.
func (c *Client) GetKYBByID(ctx context.Context, req *GetKYBByIDRequest) (*Response[GetKYBByIDResponse], error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	if req.ID == "" {
		return nil, fmt.Errorf("%w: ID", ErrMissingField)
	}

	return call[GetKYBByIDResponse](ctx, c, request{
		Method: http.MethodGet,
		Path:   "/partners/kyb/" + url.PathEscape(req.ID),
		Query:  nil,
		Body:   nil,
		Signed: true,
	})
}
