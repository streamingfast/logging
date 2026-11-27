package zapx

import (
	"unicode/utf8"

	"go.uber.org/zap"
)

const (
	// secretMask replaces the hidden part of a secret. It has a fixed length so the
	// redacted output does not reveal the length of the secret.
	secretMask = "***"

	// secretVisibleSuffix is how many trailing characters (runes) are kept visible.
	secretVisibleSuffix = 4

	// secretMinLength is the minimum length in characters (runes) the secret (type prefix
	// excluded) must have for its trailing characters (and type prefix) to be shown, this
	// ensures at least 8 characters of the secret always stay hidden.
	secretMinLength = 12

	// secretMaxPrefixSegments and secretMaxPrefixSegmentLength bound what is
	// recognized as a non-secret type prefix like `sk_live_`, `whsec_` or `ghp_`.
	secretMaxPrefixSegments      = 2
	secretMaxPrefixSegmentLength = 5
)

// SecretString is a string wrapper that redacts its content when converted to a string,
// useful for logging sensitive data like API keys, tokens or passwords while keeping
// enough visibility to tell secrets apart when debugging.
//
// The String() method prints `***` followed by the last 4 characters of the value, or
// only `***` when the value is shorter than 12 characters. An empty value is printed
// as an empty string so a missing secret stays visible.
//
// Use [SecretPrefixedString] to also keep the type prefix of known-format API keys.
//
// Example usage:
//
//	apiKey := zapx.SecretString("sk_live_...")
//	fmt.Println(apiKey) // Output: ***WXYZ
type SecretString string

func (s SecretString) String() string {
	return redactSecret(string(s))
}

// SecretPrefixedString is like [SecretString] but keeps the type prefix of known-format
// API keys visible, e.g. `sk_live_***WXYZ` for a Stripe live key.
//
// A type prefix is up to 2 segments of 1 to 5 ASCII letters each followed by `_` at the
// start of the value (`sk_live_`, `whsec_`, `ghp_`). The prefix is kept only when at
// least 12 characters remain after it, otherwise the whole value is redacted as with
// [SecretString]. An empty value is printed as an empty string.
//
// Only use it for API keys with a known prefix format, never for passwords or
// passphrases: a value made of short words joined by `_` (`blue_cat_...`) has its first
// words mistaken for a type prefix and printed in clear.
//
// Example usage:
//
//	apiKey := zapx.SecretPrefixedString("sk_live_...")
//	fmt.Println(apiKey) // Output: sk_live_***WXYZ
type SecretPrefixedString string

func (s SecretPrefixedString) String() string {
	value := string(s)

	prefix, secret := splitSecretPrefix(value)
	if prefix == "" || utf8.RuneCountInString(secret) < secretMinLength {
		return redactSecret(value)
	}

	return prefix + redactSecret(secret)
}

// Secret creates a zap field for logging secret values with redaction, see [SecretString]
// for the exact redaction rules.
//
// Example:
//
//	logger.Info("API connection", zapx.Secret("api_key", "sk_live_..."))
//	// Logs: {"api_key": "***WXYZ"}
func Secret(name string, value string) zap.Field {
	return zap.String(name, SecretString(value).String())
}

// SecretPrefixed creates a zap field for logging known-format API keys with redaction
// that keeps their type prefix visible, see [SecretPrefixedString] for the exact
// redaction rules. Never use it for passwords or passphrases, use [Secret] instead.
//
// Example:
//
//	logger.Info("API connection", zapx.SecretPrefixed("api_key", "sk_live_..."))
//	// Logs: {"api_key": "sk_live_***WXYZ"}
func SecretPrefixed(name string, value string) zap.Field {
	return zap.String(name, SecretPrefixedString(value).String())
}

func redactSecret(secret string) string {
	if secret == "" {
		return ""
	}

	if utf8.RuneCountInString(secret) < secretMinLength {
		return secretMask
	}

	// Walk back by runes so a multi-byte character is never split, which would produce
	// invalid UTF-8 and partially leak the character before the suffix.
	suffixStart := len(secret)
	for range secretVisibleSuffix {
		_, size := utf8.DecodeLastRuneInString(secret[:suffixStart])
		suffixStart -= size
	}

	return secretMask + secret[suffixStart:]
}

// splitSecretPrefix splits a recognized type prefix (e.g. `sk_live_`) from the secret
// part. When no prefix is recognized, prefix is empty and secret is the whole value.
func splitSecretPrefix(value string) (prefix string, secret string) {
	end := 0
	for range secretMaxPrefixSegments {
		segmentLength := 0
		for end+segmentLength < len(value) && isASCIILetter(value[end+segmentLength]) {
			segmentLength++
		}

		if segmentLength == 0 || segmentLength > secretMaxPrefixSegmentLength {
			break
		}

		if end+segmentLength >= len(value) || value[end+segmentLength] != '_' {
			break
		}

		end += segmentLength + 1
	}

	return value[:end], value[end:]
}

func isASCIILetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}
