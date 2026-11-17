// Onramp: create a VND deposit order, print how to pay it, read it back, then close it unless -keep is set.
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
	keep := flag.Bool("keep", false, "leave the order open so it can be paid")
	flag.Parse()

	client, err := goholdstation.NewClient(goholdstation.SandboxBaseURL, os.Getenv("HOLDSTATION_PARTNER_APP_KEY"))
	if err != nil {
		log.Fatalf("new client: %v", err)
	}

	chainID, err := strconv.ParseInt(os.Getenv("HOLDSTATION_TEST_CHAIN_ID"), 10, 64)
	if err != nil {
		log.Fatalf("HOLDSTATION_TEST_CHAIN_ID: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	referenceNo := fmt.Sprintf("sdk-sandbox-test-%d", time.Now().UnixMilli())

	created, err := client.CreateDepositOrder(ctx, &goholdstation.CreateDepositOrderRequest{
		Amount: "1000000", Currency: "vnd", ChainID: chainID,
		TokenAddress: os.Getenv("HOLDSTATION_TEST_TOKEN_ADDRESS"),
		Recipient:    os.Getenv("HOLDSTATION_TEST_RECIPIENT"),
		Callback:     os.Getenv("HOLDSTATION_TEST_CALLBACK_URL"),
		ReferenceNo:  referenceNo,
	})
	if err != nil {
		var apiErr *goholdstation.APIError
		if errors.As(err, &apiErr) {
			log.Fatalf("create deposit order refused: %s (status %d, code %s)", apiErr.Message, apiErr.StatusCode, apiErr.Code)
		}

		log.Fatalf("create deposit order: outcome unknown, look up reference %s: %v", referenceNo, err)
	}

	order := created.Data.Order
	if order.PayData != nil {
		fmt.Printf("pay %s VND to %s %s (%s)\n", order.Amount, order.PayData.BankName, order.PayData.BankAccountNumber, order.PayData.BankAccountName)
	}

	fetched, err := client.GetOrderByID(ctx, &goholdstation.GetOrderByIDRequest{OrderID: order.ID})
	if err != nil {
		log.Fatalf("get order %s: %v", order.ID, err)
	}

	fmt.Printf("order %s state %d\n", order.ID, fetched.Data.Order.State)

	if *keep {
		return
	}

	if _, err := client.CloseOrder(ctx, &goholdstation.CloseOrderRequest{OrderID: order.ID}); err != nil {
		log.Fatalf("close order %s: %v", order.ID, err)
	}

	fmt.Printf("order %s closed\n", order.ID)
}
