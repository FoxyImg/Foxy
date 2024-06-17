package utils

func Swap[T any](a T, b T) (T, T) {
	return b, a
}
