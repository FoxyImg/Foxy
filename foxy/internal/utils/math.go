package utils

func Min(ints ...int) int {
	m := ints[0]

	for _, i := range ints {
		if i < m {
			m = i
		}
	}

	return m
}

func Max(ints ...int) int {
	m := ints[0]

	for _, i := range ints {
		if i > m {
			m = i
		}
	}

	return m
}
