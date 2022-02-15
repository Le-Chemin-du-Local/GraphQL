package commerces

type CommerceErrorNotFound struct{}

func (m *CommerceErrorNotFound) Error() string {
	return "Le commerce n'a pas été trouvé"
}
