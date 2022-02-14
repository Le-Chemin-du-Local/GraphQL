package users

type UserEmailAlreadyExistsError struct{}

func (m *UserEmailAlreadyExistsError) Error() string {
	return "l'email existe déjà"
}
