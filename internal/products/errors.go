package products

type MustSpecifyCommerceIDError struct{}
type ProductNotFoundError struct{}

func (m *MustSpecifyCommerceIDError) Error() string {
	return "vous devez préciser un identifiant de commerce"
}

func (m *ProductNotFoundError) Error() string {
	return "le produit n'a pas été trouvé"
}
