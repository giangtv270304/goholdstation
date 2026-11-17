package goholdstation_test

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/giangtv270304/goholdstation"
)

const (
	sandboxTimeout = 60 * time.Second
	sandboxPrefix  = "sdk-sandbox-test-"
	sandboxKYCID   = sandboxPrefix + "kyc"
)

func sandboxContext(t *testing.T) context.Context {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), sandboxTimeout)
	t.Cleanup(cancel)

	return ctx
}

func sandboxChainID(t *testing.T, value string) int64 {
	t.Helper()

	chainID, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		t.Fatalf("%s is not an integer: %v", envTestChainID, err)
	}

	return chainID
}

func TestSandbox_ListBanks(t *testing.T) {
	t.Parallel()

	client := newSandboxClient(t)

	resp, err := client.ListBanks(sandboxContext(t), &goholdstation.ListBanksRequest{Page: 1, Size: 5})
	if err != nil {
		t.Fatalf("ListBanks: %v", err)
	}

	if len(resp.Data.Banks) == 0 || resp.Data.Banks[0].BinCode == "" {
		t.Errorf("banks = %+v, want at least one bank with a bin code", resp.Data.Banks)
	}
}

func TestSandbox_ListTokens(t *testing.T) {
	t.Parallel()

	client := newSandboxClient(t)

	resp, err := client.ListTokens(sandboxContext(t), &goholdstation.ListTokensRequest{Page: 1, PageSize: 5})
	if err != nil {
		t.Fatalf("ListTokens: %v", err)
	}

	if len(resp.Data.Tokens) == 0 || resp.Data.Tokens[0].Address == "" {
		t.Errorf("tokens = %+v, want at least one token with an address", resp.Data.Tokens)
	}
}

func TestSandbox_GetTransactionThreshold(t *testing.T) {
	t.Parallel()

	client := newSandboxClient(t)

	resp, err := client.GetTransactionThreshold(sandboxContext(t), nil)
	if err != nil {
		t.Fatalf("GetTransactionThreshold: %v", err)
	}

	if resp.Data.Config == nil {
		t.Error("config is nil")
	}
}

func TestSandbox_GetExchangeRate(t *testing.T) {
	t.Parallel()

	client := newSandboxClient(t)

	resp, err := client.GetExchangeRate(sandboxContext(t), nil)
	if err != nil {
		t.Fatalf("GetExchangeRate: %v", err)
	}

	if resp.Data.Rate == nil || resp.Data.Rate.Buy == "" || resp.Data.Rate.Sell == "" {
		t.Errorf("rate = %+v", resp.Data.Rate)
	}
}

func TestSandbox_GetExchangeRate_WrongKeyIsUnauthorized(t *testing.T) {
	t.Parallel()

	newSandboxClient(t)

	client, err := goholdstation.NewClient(sandboxBaseURL(t), "sdk-sandbox-test-wrong-key")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	_, err = client.GetExchangeRate(sandboxContext(t), nil)
	if !goholdstation.IsUnauthorized(err) {
		t.Fatalf("error = %v, want a 401 *goholdstation.APIError", err)
	}

	assertAPIError(t, err, http.StatusUnauthorized, "UNAUTHORIZED")
}

func TestSandbox_VerifyBankAccount(t *testing.T) {
	t.Parallel()

	client := newSandboxClient(t)
	env := requireEnv(t, envTestBankBinCode, envTestBankAccount)

	resp, err := client.VerifyBankAccount(sandboxContext(t), &goholdstation.VerifyBankAccountRequest{
		BinCode: env[envTestBankBinCode], AccountNumber: env[envTestBankAccount],
	})
	if err != nil {
		t.Fatalf("VerifyBankAccount: %v", err)
	}

	if resp.Data.AccountName == "" {
		t.Error("account_name is empty")
	}
}

func TestSandbox_ListOrders(t *testing.T) {
	t.Parallel()

	client := newSandboxClient(t)

	if _, err := client.ListOrders(sandboxContext(t), &goholdstation.ListOrdersRequest{Page: 1, PageSize: 1}); err != nil {
		t.Fatalf("ListOrders: %v", err)
	}
}

func TestSandbox_EstimateOrderOutcome(t *testing.T) {
	t.Parallel()

	client := newSandboxClient(t)
	env := requireEnv(t, envTestChainID, envTestTokenAddress)

	resp, err := client.EstimateOrderOutcome(sandboxContext(t), &goholdstation.EstimateOrderOutcomeRequest{
		Amount: "1000000", OrderType: goholdstation.OrderTypeOnramp,
		ChainID: sandboxChainID(t, env[envTestChainID]), TokenAddress: env[envTestTokenAddress],
	})
	if err != nil {
		t.Fatalf("EstimateOrderOutcome: %v", err)
	}

	if resp.Data.Outcome == "" || resp.Data.Rate == "" {
		t.Errorf("data = %+v", resp.Data)
	}
}

func TestSandbox_DepositOrderFlow(t *testing.T) {
	t.Parallel()

	client := newSandboxClient(t)
	env := requireEnv(t, envTestChainID, envTestTokenAddress, envTestRecipient, envTestCallbackURL)
	ctx := sandboxContext(t)

	referenceNo := fmt.Sprintf("%s%d", sandboxPrefix, time.Now().UnixNano())

	created, err := client.CreateDepositOrder(ctx, &goholdstation.CreateDepositOrderRequest{
		Amount: "1000000", Currency: "vnd",
		ChainID: sandboxChainID(t, env[envTestChainID]), TokenAddress: env[envTestTokenAddress],
		Recipient: env[envTestRecipient], Callback: env[envTestCallbackURL],
		ReferenceNo: referenceNo, KYCID: envOrEmpty(envTestKYCID),
	})
	if err != nil {
		t.Fatalf("CreateDepositOrder: %v", err)
	}

	if created.Data.Order == nil || created.Data.Order.ID == "" {
		t.Fatalf("created = %+v, want an order with an id", created.Data)
	}

	orderID := created.Data.Order.ID

	fetched, err := client.GetOrderByID(ctx, &goholdstation.GetOrderByIDRequest{OrderID: orderID})
	if err != nil {
		t.Fatalf("GetOrderByID: %v", err)
	}

	if fetched.Data.Order == nil || fetched.Data.Order.ID != orderID || fetched.Data.Order.Amount == "" {
		t.Errorf("fetched = %+v", fetched.Data.Order)
	}

	byReference, err := client.GetOrderByReferenceNo(ctx, &goholdstation.GetOrderByReferenceNoRequest{ReferenceNo: referenceNo})
	if err != nil {
		t.Fatalf("GetOrderByReferenceNo: %v", err)
	}

	if byReference.Data.Order == nil || byReference.Data.Order.ID != orderID {
		t.Errorf("by reference = %+v, want order %s", byReference.Data.Order, orderID)
	}

	if _, err := client.CloseOrder(ctx, &goholdstation.CloseOrderRequest{OrderID: orderID}); err != nil {
		t.Fatalf("CloseOrder: %v", err)
	}

	closed, err := client.GetOrderByID(ctx, &goholdstation.GetOrderByIDRequest{OrderID: orderID})
	if err != nil {
		t.Fatalf("GetOrderByID after close: %v", err)
	}

	if closed.Data.Order == nil || closed.Data.Order.State != goholdstation.OrderStateClosed {
		t.Errorf("order after close = %+v, want state %d", closed.Data.Order, goholdstation.OrderStateClosed)
	}
}

func TestSandbox_ListFunds(t *testing.T) {
	t.Parallel()

	client := newSignedSandboxClient(t)

	if _, err := client.ListFunds(sandboxContext(t), nil); err != nil {
		t.Fatalf("ListFunds: %v", err)
	}
}

func TestSandbox_ListFundLogs(t *testing.T) {
	t.Parallel()

	client := newSignedSandboxClient(t)

	if _, err := client.ListFundLogs(sandboxContext(t), &goholdstation.ListFundLogsRequest{Page: 1, PageSize: 1}); err != nil {
		t.Fatalf("ListFundLogs: %v", err)
	}
}

func TestSandbox_ListFundDepositAddresses(t *testing.T) {
	t.Parallel()

	client := newSignedSandboxClient(t)

	if _, err := client.ListFundDepositAddresses(sandboxContext(t), nil); err != nil {
		t.Fatalf("ListFundDepositAddresses: %v", err)
	}
}

func TestSandbox_ListDepositVAs(t *testing.T) {
	t.Parallel()

	client := newSignedSandboxClient(t)

	if _, err := client.ListDepositVAs(sandboxContext(t), nil); err != nil {
		t.Fatalf("ListDepositVAs: %v", err)
	}
}

func TestSandbox_ListMIDVAs(t *testing.T) {
	t.Parallel()

	client := newSignedSandboxClient(t)

	if _, err := client.ListMIDVAs(sandboxContext(t), nil); err != nil {
		t.Fatalf("ListMIDVAs: %v", err)
	}
}

func TestSandbox_GetWebhookURL(t *testing.T) {
	t.Parallel()

	client := newSignedSandboxClient(t)

	if _, err := client.GetWebhookURL(sandboxContext(t), nil); err != nil {
		t.Fatalf("GetWebhookURL: %v", err)
	}
}

func TestSandbox_ListKYBs(t *testing.T) {
	t.Parallel()

	client := newSignedSandboxClient(t)

	if _, err := client.ListKYBs(sandboxContext(t), &goholdstation.ListKYBsRequest{Page: 1, PageSize: 1}); err != nil {
		t.Fatalf("ListKYBs: %v", err)
	}
}

func TestSandbox_ListKYCs(t *testing.T) {
	t.Parallel()

	client := newSignedSandboxClient(t)

	if _, err := client.ListKYCs(sandboxContext(t), &goholdstation.ListKYCsRequest{Page: 1, PageSize: 1}); err != nil {
		t.Fatalf("ListKYCs: %v", err)
	}
}

func TestSandbox_KYCFlow(t *testing.T) {
	t.Parallel()

	client := newSignedSandboxClient(t)
	ctx := sandboxContext(t)

	_, err := client.GetKYCByID(ctx, &goholdstation.GetKYCByIDRequest{ID: sandboxKYCID})
	if goholdstation.IsNotFound(err) {
		_, err = client.CreateKYC(ctx, &goholdstation.CreateKYCRequest{
			ID: sandboxKYCID,
			KYC: goholdstation.KYCInput{
				Name: "SDK Sandbox Test", Email: "sdk-sandbox-test@example.com", Phone: "+84900000000",
				Address: "1 Test Street", IDNumber: "000000000000",
				IDType: goholdstation.KYCIDTypeCitizenID, Sex: goholdstation.KYCSexOther, Nationality: "VN",
				DateOfIssue: "2020-01-01", DateOfExpiry: "2035-01-01",
			},
		})
		if err != nil {
			t.Fatalf("CreateKYC: %v", err)
		}

		_, err = client.GetKYCByID(ctx, &goholdstation.GetKYCByIDRequest{ID: sandboxKYCID})
	}

	if err != nil {
		t.Fatalf("GetKYCByID: %v", err)
	}
}

func TestSandbox_GetKYCByID_NotFound(t *testing.T) {
	t.Parallel()

	client := newSignedSandboxClient(t)

	missingID := fmt.Sprintf("%smissing-%d", sandboxPrefix, time.Now().UnixNano())

	_, err := client.GetKYCByID(sandboxContext(t), &goholdstation.GetKYCByIDRequest{ID: missingID})
	if !goholdstation.IsNotFound(err) {
		t.Fatalf("error = %v, want a 404 *goholdstation.APIError", err)
	}

	assertAPIError(t, err, http.StatusNotFound, "KYC_NOT_FOUND")
}
