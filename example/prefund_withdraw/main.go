// Prefund withdraw: quote a VND payout to a bank account from the prefunded balance.
// The payout is confirmed only with -confirm, because confirming moves money.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/giangtv270304/goholdstation"
)

func main() {
	confirm := flag.Bool("confirm", false, "confirm the quote and send the payout")
	bankID := flag.Int64("bank-id", 0, "Bank.ID from ListBanks")
	fullName := flag.String("full-name", "", "account holder name")
	token := flag.String("token", "USDT", "prefunded token to debit")
	flag.Parse()

	client, err := goholdstation.NewClient(
		goholdstation.SandboxBaseURL,
		os.Getenv("HOLDSTATION_PARTNER_APP_KEY"),
		goholdstation.WithPrivateKey(os.Getenv("HOLDSTATION_ED25519_PRIVATE_KEY")),
	)
	if err != nil {
		log.Fatalf("new client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	idempotencyKey := "sdk-sandbox-test-" + strconv.FormatInt(time.Now().UnixMilli(), 10)

	quoted, err := client.CreateWithdrawOrder(ctx, &goholdstation.CreateWithdrawOrderRequest{
		Amount: "100000", Currency: "vnd", TokenSymbol: *token,
		PaymentInfo: &goholdstation.PaymentInfo{
			BankID: *bankID, FullName: *fullName, AccountType: 1,
			AccountNumber: os.Getenv("HOLDSTATION_TEST_BANK_ACCOUNT_NUMBER"),
		},
		Callback:       os.Getenv("HOLDSTATION_TEST_CALLBACK_URL"),
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		var apiErr *goholdstation.APIError
		if errors.As(err, &apiErr) {
			log.Fatalf("create withdraw order refused: %s (status %d, code %s)", apiErr.Message, apiErr.StatusCode, apiErr.Code)
		}

		log.Fatalf("create withdraw order: outcome unknown, resend with idempotency key %s: %v", idempotencyKey, err)
	}

	order := quoted.Data.Order
	fmt.Printf("quote %s: %s %s for %s %s at rate %s\n", order.ID, order.Amount, order.Currency, order.Outcome, order.OutcomeCurrency, order.Rate)

	if !*confirm {
		fmt.Println("not confirmed; run with -confirm to send the payout")

		return
	}

	confirmed, err := client.ConfirmWithdrawOrder(ctx, &goholdstation.ConfirmWithdrawOrderRequest{OrderID: order.ID})
	if err != nil {
		log.Fatalf("confirm order %s: outcome unknown unless this is a 4xx APIError, read the order before retrying: %v", order.ID, err)
	}

	fmt.Printf("order %s confirmed, state %d\n", order.ID, confirmed.Data.Order.State)
}
