package helpers

func First[T any](arr []T, byKey func(T) bool) *T {
	for _, el := range arr {
		if byKey(el) {
			return &el
		}
	}
	return nil
}

func ToPtr[T any](v T) *T {
	return &v
}

// TODO: use reflection as query
func GetFromArr[T any](arr []T, q func (T) bool) *T  {
	for _, v := range arr {
		if q(v) {
			return &v
		}
	}
	return nil
}