package utils

func Contains[T comparable](arr []T, item T) bool {
	for _, v := range arr {
		if item == v {
			return true
		}
	}
	return false
}
