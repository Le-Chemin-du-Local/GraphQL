package commands

type MustSpecifyCommerceIDError struct{}
type CommerceCommandNotFoundError struct{}
type CommandNotFoundError struct{}

func (m *MustSpecifyCommerceIDError) Error() string {
	return "vous devez préciser un identifiant de commerce"
}

func (m *CommerceCommandNotFoundError) Error() string {
	return "La commande correspondant au commerce n'a pas été trouvée"
}

func (m *CommandNotFoundError) Error() string {
	return "La commande n'a pas été trouvée"
}
