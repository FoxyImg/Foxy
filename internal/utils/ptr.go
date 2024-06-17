package utils

func Ptr[T any](v T) *T {
	return &v
}

func IfNil[T any](v *T, defaultValue T) T {
	if v == nil {
		return defaultValue
	} else {
		return *v
	}
}

func IfNiPtr[T any](v *T, defaultValue T) *T {
	if v == nil {
		return &defaultValue
	} else {
		return v
	}
}
