package stripehandler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"chemin-du-local.bzh/graphql/internal/config"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/paymentintent"
)

type item struct {
	id string
}

func calculateOrderAmount(items []item) int64 {
	return 4000
}

func HandleCreatePaymentIntent(w http.ResponseWriter, r *http.Request) {
	// Set Stripe API key
	apiKey := config.Cfg.Stripe.Key
	stripe.Key = apiKey

	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UseStripeSDK    *bool   `json:"useStripeSKD"`
		PaymentMethodID *string `json:"paymentMethodId"`
		PaymentIntentID *string `json:"paymentIntentId"`
		Items           []item  `json:"items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("json.NewDecoder.Decode: %v", err)
		return
	}

	fmt.Println(req)
	if req.PaymentIntentID != nil {
		params := &stripe.PaymentIntentConfirmParams{
			PaymentMethod: req.PaymentMethodID,
		}

		pi, err := paymentintent.Confirm(*req.PaymentIntentID, params)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, struct {
			ClientSecret string `json:"clientSecret"`
		}{
			ClientSecret: pi.ClientSecret,
		})
	} else if req.PaymentMethodID != nil {

		// Create a PaymentIntent with amount and currency
		confirm := true
		confirmationMethode := "manual"
		params := &stripe.PaymentIntentParams{
			UseStripeSDK:       req.UseStripeSDK,
			Amount:             stripe.Int64(calculateOrderAmount(req.Items)),
			Currency:           stripe.String(string(stripe.CurrencyEUR)),
			PaymentMethod:      req.PaymentMethodID,
			Confirm:            &confirm,
			ConfirmationMethod: &confirmationMethode,
		}

		pi, err := paymentintent.New(params)
		log.Printf("pi.New: %v", pi.ClientSecret)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Printf("pi.New: %v", err)
			return
		}

		writeJSON(w, struct {
			ClientSecret   string `json:"clientSecret"`
			RequiresAction bool   `json:"requiresAction"`
		}{
			ClientSecret:   pi.ClientSecret,
			RequiresAction: pi.Status == stripe.PaymentIntentStatusRequiresAction,
		})
	}
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("json.NewEncoder.Encode: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := io.Copy(w, &buf); err != nil {
		log.Printf("io.Copy: %v", err)
		return
	}
}
