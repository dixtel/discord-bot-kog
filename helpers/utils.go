package helpers

func First[T any](arr []T, byKey func(T) bool) *T {
	for _, el := range arr {
		if byKey(el) {
			return &el
		}
	}
	return nil
}
