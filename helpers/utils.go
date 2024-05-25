package helpers

import (
	"reflect"
	"runtime"
)

func First[T any](arr []T, byKey func(T) bool) (ret T) {
	for _, el := range arr {
		if byKey(el) {
			return el
		}
	}

	return ret
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

func GetFunctionName(i interface{}) string {
    return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}