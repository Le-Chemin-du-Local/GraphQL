package users

type UserAccessDenied struct{}
type UserEmailAlreadyExistsError struct{}
type UserPasswordIncorrect struct{}

func (m *UserAccessDenied) Error() string {
	return "access denied"
}

func (m *UserEmailAlreadyExistsError) Error() string {
	return "l'email existe déjà"
}

func (m *UserPasswordIncorrect) Error() string {
	return "le mot de passe n'est pas correcte"
}
