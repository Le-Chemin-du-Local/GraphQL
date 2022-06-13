package registeredpaymentmethod

import (
	"chemin-du-local.bzh/graphql/graph/model"
	"github.com/stripe/stripe-go/v72/paymentmethod"
)

type RegisteredPaymentMethod struct {
	Name     string `bson:"name"`
	StripeID string `bson:"stripeID"`
}

func GetPaymentMethodDetails(paymentMethodId string) (*model.RegisteredPaymentMethod, error) {
	paymentMethod, err := paymentmethod.Get(paymentMethodId, nil)

	if err != nil {
		return nil, err
	}

	if paymentMethod == nil {
		return nil, &PaymentMethodNotFoundError{}
	}

	return &model.RegisteredPaymentMethod{
		StripeID:        paymentMethodId,
		CardBrand:       (*string)(&paymentMethod.Card.Brand), // TODO: hmm
		CardLast4Digits: &paymentMethod.Card.Last4,
	}, nil
}
