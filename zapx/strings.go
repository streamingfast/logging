package zapx

import (
	"fmt"

	"go.uber.org/zap"
)

var (
	// StringKeyf is an alias for [Stringkf] using a longer, more consistent name.
	//
	// See [Stringkf] for more details.
	StringKeyf = Stringkf

	// StringerKeyf is an alias for [Stringerkf] using a longer, more consistent name.
	//
	// See [Stringerkf] for more details.
	StringerKeyf = Stringerkf
)

// Stringkf creates a zap field with a string value and a key formatted using
// fmt.Sprintf.
//
// It's important to note that the key formatting is performed immediately when
// the field is created, not deferred. This means that any computation involved
// in formatting the key will occur even if the field is not logged.
//
// Important: This is not recommended for high-frequency trace logs as the
// formatting is done eagerly.
//
// Example:
//
//	zlog.Info("User logged in", zapx.Stringkf("welcome_message_%d", value, userID))
//
//go:inline
func Stringkf(format string, value string, args ...any) zap.Field {
	return zap.String(fmt.Sprintf(format, args...), value)
}

// Stringerkf creates a zap field with a fmt.Stringer value and a key formatted
// using fmt.Sprintf.
//
// It's important to note that the key formatting is performed immediately when
// the field is created, not deferred. This means that any computation involved
// in formatting the key will occur even if the field is not logged.
//
// Important: This is not recommended for high-frequency trace logs as the
// formatting is done eagerly.
//
// Example:
//
//	zlog.Info("User logged in", zapx.Stringerkf("user_%d", userStringer, userID))
//
//go:inline
func Stringerkf(format string, stringer fmt.Stringer, args ...any) zap.Field {
	return zap.Stringer(fmt.Sprintf(format, args...), stringer)
}

// Stringf creates a zap field with a string value formatted using fmt.Sprintf.
// The formatting is deferred until the field is actually serialized, which can
// save computation if the field is not logged.
//
// Example:
//
//	zlog.Info("User logged in", zapx.Stringf("welcome_message", "Hello, %s!", userName))
//
//go:inline
func Stringf(name string, valueFormat string, args ...any) zap.Field {
	return zap.Stringer(name, deferredSprintf{format: valueFormat, args: args})
}

type deferredSprintf struct {
	format string
	args   []any
}

func (d deferredSprintf) String() string {
	return fmt.Sprintf(d.format, d.args...)
}
