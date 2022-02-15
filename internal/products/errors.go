package products

type MustSpecifyCommerceIDError struct{}

func (m *MustSpecifyCommerceIDError) Error() string {
	return "vous devez préciser un identifiant de commerce"
}
