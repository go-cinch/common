package utils

// Deprecated: use github.com/samber/lo Contains instead
func Contains[T comparable](arr []T, item T) bool {
	for _, v := range arr {
		if item == v {
			return true
		}
	}
	return false
}
