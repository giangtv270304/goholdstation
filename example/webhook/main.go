// Webhook: verify Holdstation webhooks with the checksum key, then handle order and fund events.
package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/giangtv270304/goholdstation"
)

func main() {
	checksumKey := os.Getenv("HOLDSTATION_WEBHOOK_CHECKSUM_KEY")

	http.HandleFunc("/webhooks/holdstation", func(w http.ResponseWriter, r *http.Request) {
		handle(w, r, checksumKey)
	})

	server := &http.Server{Addr: ":8080", ReadHeaderTimeout: 10 * time.Second}
	log.Fatal(server.ListenAndServe())
}

func handle(w http.ResponseWriter, r *http.Request, checksumKey string) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		http.Error(w, "body too large", http.StatusRequestEntityTooLarge)

		return
	}

	signature := r.Header.Get(goholdstation.WebhookSignatureHeader)
	if err := goholdstation.VerifyWebhookSignature(body, signature, checksumKey); err != nil {
		http.Error(w, "bad signature", http.StatusUnauthorized)

		return
	}

	event, err := goholdstation.ParseWebhookEvent(body)
	if err != nil {
		http.Error(w, "bad body", http.StatusBadRequest)

		return
	}

	switch event.Topic {
	case goholdstation.WebhookTopicOrderStatusUpdated:
		payload, err := event.OrderStatusUpdated()
		if err != nil {
			http.Error(w, "bad payload", http.StatusBadRequest)

			return
		}

		log.Printf("event %s: order %s state %d -> %d", event.ID, payload.OrderID, payload.OldOrderState, payload.NewOrderState)
	case goholdstation.WebhookTopicFundBalanceChanged:
		payload, err := event.FundBalanceChanged()
		if err != nil {
			http.Error(w, "bad payload", http.StatusBadRequest)

			return
		}

		log.Printf("event %s: fund %s balance changed from %s", event.ID, payload.Token, payload.PreBalance)
	default:
		log.Printf("event %s: topic %s", event.ID, event.Topic)
	}

	w.WriteHeader(http.StatusOK)
}
