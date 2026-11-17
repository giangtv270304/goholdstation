package goholdstation_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/giangtv270304/goholdstation"
)

const kycJSON = `{"id":"cust 1/a","partner_id":"partner-xyz","name":"NGUYEN VAN A","email":"a@example.com","phone":"+84900000000","address":"HCMC","id_number":"079090000001","id_type":1,"sex":1,"nationality":"VN","date_of_issue":"2021-05-06","date_of_expiry":"2031-05-06","created_at":{"seconds":1700000000,"nanos":0},"updated_at":{"seconds":1700000001,"nanos":0},"id_front_img":{"file_ref":"f_front","content_type":"image/jpeg","size":20480}}`

var kycBodies = map[string]string{
	"data.kyc":    `{"message":"success","data":{"success":true,"kyc":` + kycJSON + `}}`,
	"flat record": `{"message":"success","data":` + kycJSON + `}`,
}

func assertKYC(t *testing.T, kyc *goholdstation.KYC) {
	t.Helper()

	if kyc == nil || kyc.ID != "cust 1/a" || kyc.PartnerID != "partner-xyz" || kyc.Name != "NGUYEN VAN A" || kyc.IDNumber != "079090000001" {
		t.Fatalf("kyc = %+v", kyc)
	}

	if kyc.IDType != goholdstation.KYCIDTypeCitizenID || kyc.Sex != goholdstation.KYCSexMale || kyc.DateOfIssue != "2021-05-06" || kyc.DateOfExpiry != "2031-05-06" {
		t.Errorf("kyc = %+v", kyc)
	}

	if kyc.CreatedAt == nil || kyc.CreatedAt.Seconds != 1700000000 || kyc.UpdatedAt == nil || kyc.UpdatedAt.Seconds != 1700000001 {
		t.Errorf("timestamps = %+v %+v", kyc.CreatedAt, kyc.UpdatedAt)
	}

	if kyc.IDFrontImg == nil || kyc.IDFrontImg.FileRef != "f_front" || kyc.IDFrontImg.Size != 20480 || kyc.IDBackImg != nil || kyc.SelfieImg != nil {
		t.Errorf("images = %+v %+v %+v", kyc.IDFrontImg, kyc.IDBackImg, kyc.SelfieImg)
	}
}

func TestCreateKYC_Success_BothDocumentedForms(t *testing.T) {
	t.Parallel()

	for name, body := range kycBodies {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client, api := newFakeAPI(t, http.StatusOK, body)

			resp, err := client.CreateKYC(context.Background(), &goholdstation.CreateKYCRequest{
				ID: "cust 1/a",
				KYC: goholdstation.KYCInput{
					Name: "NGUYEN VAN A", Email: "a@example.com", Phone: "+84900000000", Address: "HCMC", IDNumber: "079090000001",
					IDType: goholdstation.KYCIDTypeCitizenID, Sex: goholdstation.KYCSexMale, Nationality: "VN",
					DateOfIssue: "2021-05-06", DateOfExpiry: "2031-05-06",
					IDFrontImg: &goholdstation.KYCImageInput{ContentType: "image/jpeg", Data: "ZnJvbnQ="},
				},
			})
			if err != nil {
				t.Fatalf("CreateKYC: %v", err)
			}

			api.assertRequest(t, wantRequest{
				method: http.MethodPost, path: "/partners/kycs/cust%201%2Fa", query: "",
				body:   `{"name":"NGUYEN VAN A","email":"a@example.com","phone":"+84900000000","address":"HCMC","id_number":"079090000001","id_type":1,"sex":1,"nationality":"VN","date_of_issue":"2021-05-06","date_of_expiry":"2031-05-06","id_front_img":{"content_type":"image/jpeg","data":"ZnJvbnQ="}}`,
				signed: true,
			})

			assertKYC(t, resp.Data.KYC)
		})
	}
}

func TestCreateKYC_APIError(t *testing.T) {
	t.Parallel()

	client, _ := newFakeAPI(t, http.StatusConflict, `{"error":"conflict","error_code":"KYC_ALREADY_EXISTS","message":"KYC already exists"}`)

	_, err := client.CreateKYC(context.Background(), &goholdstation.CreateKYCRequest{ID: "cust-1"})
	if !goholdstation.IsConflict(err) {
		t.Errorf("error = %v, want IsConflict", err)
	}

	assertAPIError(t, err, http.StatusConflict, "KYC_ALREADY_EXISTS")
}
