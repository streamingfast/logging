package zapx

import (
	"reflect"

	"go.uber.org/zap"
)

type typeStringer struct {
	inner any
}

func (s typeStringer) String() string {
	if s.inner == nil {
		return "<nil>"
	}

	return reflect.TypeOf(s.inner).String()
}

// Type returns a zap.Field that for which is going to render
// the field "<name>": "<type name of value>" where `<type name of value>` is
// going to be the `reflect.TypeOf(value).String()`.
//
// The field is lazy and will not be evaluated until the field is actually
// logged, which is useful for performance reasons.
func Type(name string, value any) zap.Field {
	return zap.Stringer(name, typeStringer{inner: value})
}
