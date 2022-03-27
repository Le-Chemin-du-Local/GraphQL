package paniers

type PanierNotFoundError struct{}

func (m *PanierNotFoundError) Error() string {
	return "le panier n'a pas été trouvé"
}
