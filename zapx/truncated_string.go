package zapx

import (
	"fmt"

	"go.uber.org/zap"
)

type truncatedStringField struct {
	value     fmt.Stringer
	maxLength int
}

func (t truncatedStringField) String() string {
	value := t.value.String()
	if len(value) <= t.maxLength {
		return value
	}
	return value[:t.maxLength]
}

// TruncatedString creates a zap.Field that truncates a string value to a maximum length.
// The truncation is performed lazily when the field is actually logged, using zap's
// lazy loading mechanism through the String() method.
//
// If the string is shorter than or equal to maxLength, it is returned as-is.
// If the string is longer than maxLength, it is truncated to maxLength characters.
//
// Example usage:
//
//	logger.Info("Processing data", zapx.TruncatedString("data", longString, 100))
func TruncatedString(fieldName string, value string, maxLength int) zap.Field {
	return zap.Stringer(fieldName, truncatedStringField{
		value:     stringer(value),
		maxLength: maxLength,
	})
}

// TruncatedStringer creates a zap.Field that truncates a fmt.Stringer value to a maximum length.
// The truncation is performed lazily when the field is actually logged, using zap's
// lazy loading mechanism through the String() method.
//
// If the string representation of the fmt.Stringer is shorter than or equal to maxLength,
// it is returned as-is. If it is longer than maxLength, it is truncated to maxLength characters.
//
// Example usage:
//
//	logger.Info("Processing data", zapx.TruncatedStringer("data", longStringer, 100))
func TruncatedStringer(fieldName string, stringer fmt.Stringer, maxLength int) zap.Field {
	return zap.Stringer(fieldName, truncatedStringField{
		value:     stringer,
		maxLength: maxLength,
	})
}

type stringer string

func (f stringer) String() string {
	return string(f)
}
