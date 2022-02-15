package commerces

type CommerceErrorNotFound struct{}
type NoCommerceForUserError struct{}

func (m *CommerceErrorNotFound) Error() string {
	return "Le commerce n'a pas été trouvé"
}

func (m *NoCommerceForUserError) Error() string {
	return "aucun commerce trouvé pour l'utilisateur"
}
