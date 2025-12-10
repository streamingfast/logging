package zapx

import (
	"math"
	"strings"

	"go.uber.org/zap"
)

// SecretString is a string wrapper that redacts most of its content when converted to a string,
// showing only a portion at the beginning and end. This is useful for logging sensitive data
// like API keys, tokens, or passwords while maintaining some visibility for debugging.
//
// The String() method shows approximately 25% of characters from the start and 10% from the end
// (minimum 1 character at the end), with the middle replaced by asterisks.
//
// Example usage:
//
//	apiKey := zapx.SecretString("ac_fake_abcdefghijklmnopqrstuvwxyz123456")
//	fmt.Println(apiKey) // Output: ac_fake_abc***3456
type SecretString string

func (s SecretString) String() string {
	length := len(s)
	if length <= 3 {
		return strings.Repeat("*", length)
	}

	// Show 25% at the front, minimum 1
	frontCount := int(math.Max(1, math.Floor(0.25*float64(length))))

	// Show 10% at the end, minimum 1
	endCount := int(math.Max(1, math.Floor(0.10*float64(length))))

	// Ensure we don't show more than the string length
	if frontCount+endCount >= length {
		return strings.Repeat("*", length)
	}

	front := string(s[:frontCount])
	end := string(s[length-endCount:])
	middleCount := length - frontCount - endCount

	return front + strings.Repeat("*", middleCount) + end
}

// Secret creates a zap field for logging secret values with redaction.
// It shows approximately 25% of characters from the start and 10% from the end,
// with the middle replaced by asterisks.
//
// Example:
//
//	logger.Info("API connection", zapx.Secret("api_key", "ac_fake_abcdefghijklmnopqrstuvwxyz123456"))
//	// Logs: {"api_key": "ac_fake_abc***3456"}
func Secret(name string, value string) zap.Field {
	return zap.String(name, SecretString(value).String())
}
