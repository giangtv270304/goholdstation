// Quickstart: read the USD/VND rate, then the prefunded balances, in the sandbox.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/giangtv270304/goholdstation"
)

func main() {
	client, err := goholdstation.NewClient(
		goholdstation.SandboxBaseURL,
		os.Getenv("HOLDSTATION_PARTNER_APP_KEY"),
		goholdstation.WithPrivateKey(os.Getenv("HOLDSTATION_ED25519_PRIVATE_KEY")),
	)
	if err != nil {
		log.Fatalf("new client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rate, err := client.GetExchangeRate(ctx, nil)
	if err != nil {
		var apiErr *goholdstation.APIError
		if errors.As(err, &apiErr) {
			log.Fatalf("get exchange rate: %s (status %d, code %s)", apiErr.Message, apiErr.StatusCode, apiErr.Code)
		}

		log.Fatalf("get exchange rate: %v", err)
	}

	if rate.Data.Rate != nil {
		fmt.Printf("USD/VND buy %s sell %s\n", rate.Data.Rate.Buy, rate.Data.Rate.Sell)
	}

	funds, err := client.ListFunds(ctx, nil)
	if errors.Is(err, goholdstation.ErrMissingPrivateKey) {
		fmt.Println("set HOLDSTATION_ED25519_PRIVATE_KEY to call the signed operations")

		return
	}

	if err != nil {
		log.Fatalf("list funds: %v", err)
	}

	for _, fund := range funds.Data.Funds {
		fmt.Printf("%s %s\n", fund.Token, fund.Balance)
	}
}
