package clickandcollect

type CCCommandNotFoundError struct{}

func (m *CCCommandNotFoundError) Error() string {
	return "la command n'a pas été trouvée"
}
