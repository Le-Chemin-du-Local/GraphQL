package registeredpaymentmethod

type PaymentMethodNotFoundError struct{}

func (m *PaymentMethodNotFoundError) Error() string {
	return "la methode de paiement n existe pas"
}
