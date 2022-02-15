package commerces

type CommerceErrorNotFound struct{}

func (m *CommerceErrorNotFound) Error() string {
	return "the commerce is not found"
}
