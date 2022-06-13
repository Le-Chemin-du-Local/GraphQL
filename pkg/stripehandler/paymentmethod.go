package stripehandler

import (
	"fmt"

	"chemin-du-local.bzh/graphql/graph/model"
	"chemin-du-local.bzh/graphql/internal/config"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/paymentmethod"
)

func GetPaymentMethods(customerID string) []*model.RegisteredPaymentMethod {
	// Set Stripe API key
	apiKey := config.Cfg.Stripe.Key
	stripe.Key = apiKey

	result := []*model.RegisteredPaymentMethod{}

	params := &stripe.PaymentMethodListParams{
		Customer: stripe.String(customerID),
		Type:     stripe.String("card"),
	}
	i := paymentmethod.List(params)
	fmt.Println(i)
	for i.Next() {
		pm := i.PaymentMethod()

		result = append(result, &model.RegisteredPaymentMethod{
			Name:            pm.Metadata["name"],
			StripeID:        pm.ID,
			CardBrand:       (*string)(&pm.Card.Brand), // TODO: hmm
			CardLast4Digits: &pm.Card.Last4,
		})
	}

	return result
}
