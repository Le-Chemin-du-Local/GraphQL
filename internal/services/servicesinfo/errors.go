package servicesinfo

type ServiceNotFoundError struct{}

func (m *ServiceNotFoundError) Error() string {
	return "le service n'a pas été trouvé"
}
