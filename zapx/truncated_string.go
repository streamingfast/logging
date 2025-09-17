package zapx

import (
	"go.uber.org/zap"
)

type truncatedStringField struct {
	value     string
	maxLength int
}

func (t truncatedStringField) String() string {
	if len(t.value) <= t.maxLength {
		return t.value
	}
	return t.value[:t.maxLength]
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
		value:     value,
		maxLength: maxLength,
	})
}
