package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

const kybJSON = `{"id":"kyb_01HQ","partner_id":"partner-xyz","subject_type":2,"reference_id":"merchant-7","legal_name":"SDK Shop Co","registration_number":"0312345678","address":"1 Le Loi, HCMC","nature_of_business":"retail","incorporation_country":"VN","incorporation_date":"2020-01-15","email":"ops@example.com","phone":"+84900000000","website":"https://shop.example","geographic_scope":["VN","SG"],"persons":[{"role":1,"name":"NGUYEN VAN A","date_of_birth":"1990-02-03","address":"HCMC","id_number":"079090000001","id_type":1,"sex":1,"nationality":"VN","date_of_issue":"2021-05-06","documents":[{"type":11,"files":[{"file_ref":"f_0123456789ab","filename":"id.jpg","content_type":"image/jpeg","size":20480}],"date_of_issue":"2021-05-06"}],"selfie":{"file_ref":"f_0123456789ac","filename":"selfie.png","content_type":"image/png","size":1024}}],"documents":[{"type":1,"files":[{"file_ref":"f_0123456789ad","filename":"inc.pdf","content_type":"application/pdf","size":4096}],"date_of_issue":"2020-01-15"},{"type":4,"files":[],"document_number":"LIC-9","date_of_issue":"2022-01-01","date_of_expiry":"2027-01-01"}],"status":1,"rejection_reason":"","submitted_at":"2026-10-29T03:14:22Z","created_at":"2026-10-29T03:14:22Z","updated_at":{"seconds":1793243662,"nanos":0},"reviewed_at":null}`

func TestSubmitKYB_Success(t *testing.T) {
	t.Parallel()

	client, api := newFakeAPI(t, http.StatusCreated, `{"message":"success","data":{"success":true,"kyb":`+kybJSON+`}}`)

	resp, err := client.SubmitKYB(context.Background(), &goholdstation.SubmitKYBRequest{
		SubjectType: goholdstation.KYBSubjectTypeSubMerchant, ReferenceID: "merchant-7",
		LegalName: "SDK Shop Co", RegistrationNumber: "0312345678", Address: "1 Le Loi, HCMC", NatureOfBusiness: "retail",
		IncorporationCountry: "VN", IncorporationDate: "2020-01-15", Email: "ops@example.com",
		GeographicScope: []string{"VN", "SG"},
		Persons: []goholdstation.KYBPersonInput{{
			Role: goholdstation.KYBPersonRoleLegalRepresentative, Name: "NGUYEN VAN A", DateOfBirth: "1990-02-03", Address: "HCMC",
			IDNumber: "079090000001", IDType: 1, Sex: 1, Nationality: "VN", DateOfIssue: "2021-05-06",
			Documents: []goholdstation.KYBDocumentInput{{
				Type:  goholdstation.KYBDocumentTypeCitizenID,
				Files: []goholdstation.KYBFileInput{{Filename: "id.jpg", ContentType: "image/jpeg", Data: "aWQ="}},
			}},
			Selfie: &goholdstation.KYBFileInput{Filename: "selfie.png", ContentType: "image/png", Data: "c2VsZmll"},
		}},
		Documents: []goholdstation.KYBDocumentInput{
			{Type: goholdstation.KYBDocumentTypeIncorporationPaper, Files: []goholdstation.KYBFileInput{{Filename: "inc.pdf", ContentType: "application/pdf", Data: "cGRm"}}, DateOfIssue: "2020-01-15"},
			{Type: goholdstation.KYBDocumentTypeImportExportLicence, Files: []goholdstation.KYBFileInput{}, DocumentNumber: "LIC-9", DateOfIssue: "2022-01-01", DateOfExpiry: "2027-01-01"},
		},
	})
	if err != nil {
		t.Fatalf("SubmitKYB: %v", err)
	}

	api.assertRequest(t, wantRequest{
		method: http.MethodPost, path: "/partners/kyb", query: "",
		body:   `{"subject_type":2,"reference_id":"merchant-7","legal_name":"SDK Shop Co","registration_number":"0312345678","address":"1 Le Loi, HCMC","nature_of_business":"retail","incorporation_country":"VN","incorporation_date":"2020-01-15","email":"ops@example.com","geographic_scope":["VN","SG"],"persons":[{"role":1,"name":"NGUYEN VAN A","date_of_birth":"1990-02-03","address":"HCMC","id_number":"079090000001","id_type":1,"sex":1,"nationality":"VN","date_of_issue":"2021-05-06","documents":[{"type":11,"files":[{"filename":"id.jpg","content_type":"image/jpeg","data":"aWQ="}]}],"selfie":{"filename":"selfie.png","content_type":"image/png","data":"c2VsZmll"}}],"documents":[{"type":1,"files":[{"filename":"inc.pdf","content_type":"application/pdf","data":"cGRm"}],"date_of_issue":"2020-01-15"},{"type":4,"files":[],"document_number":"LIC-9","date_of_issue":"2022-01-01","date_of_expiry":"2027-01-01"}]}`,
		signed: true,
	})

	assertKYB(t, resp.Data.KYB)
}

func assertKYB(t *testing.T, kyb *goholdstation.KYB) {
	t.Helper()

	if kyb == nil || kyb.ID != "kyb_01HQ" || kyb.PartnerID != "partner-xyz" || kyb.SubjectType != 2 || kyb.ReferenceID != "merchant-7" || kyb.LegalName != "SDK Shop Co" {
		t.Fatalf("kyb = %+v", kyb)
	}

	if kyb.Status != goholdstation.KYBStatusPending || kyb.RejectionReason != "" || len(kyb.GeographicScope) != 2 || kyb.IncorporationDate != "2020-01-15" {
		t.Errorf("kyb = %+v", kyb)
	}

	if kyb.SubmittedAt == nil || kyb.SubmittedAt.Seconds != 1793243662 || kyb.UpdatedAt == nil || kyb.UpdatedAt.Seconds != 1793243662 || kyb.ReviewedAt != nil {
		t.Errorf("timestamps = %+v %+v %+v", kyb.SubmittedAt, kyb.UpdatedAt, kyb.ReviewedAt)
	}

	if len(kyb.Persons) != 1 || len(kyb.Persons[0].Documents) != 1 || kyb.Persons[0].Selfie == nil || kyb.Persons[0].IDCardFront != nil {
		t.Fatalf("persons = %+v", kyb.Persons)
	}

	file := kyb.Persons[0].Documents[0].Files[0]
	if kyb.Persons[0].Documents[0].Type != goholdstation.KYBDocumentTypeCitizenID || file.FileRef != "f_0123456789ab" || file.Filename != "id.jpg" || file.Size != 20480 {
		t.Errorf("person document = %+v", kyb.Persons[0].Documents[0])
	}

	if len(kyb.Documents) != 2 || kyb.Documents[1].DocumentNumber != "LIC-9" || kyb.Documents[1].DateOfExpiry != "2027-01-01" || len(kyb.Documents[1].Files) != 0 {
		t.Errorf("documents = %+v", kyb.Documents)
	}
}

func TestSubmitKYB_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusConflict, `{"error":"CONFLICT","message":"A pending KYB record already exists"}`)

	_, err := client.SubmitKYB(context.Background(), &goholdstation.SubmitKYBRequest{SubjectType: goholdstation.KYBSubjectTypePartner})
	if !goholdstation.IsConflict(err) {
		t.Errorf("error = %v, want IsConflict", err)
	}

	assertAPIError(t, err, http.StatusConflict, "CONFLICT")
}
