package utils

func DefaultIfZero[T comparable](value T, defaultValue T) T {
	var zero T // zero is automatically initialized to the zero value of type T
	if value == zero {
		return defaultValue
	}
	return value
}
