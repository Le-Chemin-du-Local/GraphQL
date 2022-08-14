package utils

func RemoveString(slice []string, index int) []string {
	return append(slice[:index], slice[index+1:]...)
}
