package zapx

import (
	"math"
	"strings"

	"go.uber.org/zap"
)

type SecretString string

func (s SecretString) String() string {
	if len(s) <= 3 {
		return strings.Repeat("*", len(s))
	}

	count := len(s)
	base := int(math.Floor(1.0 / 4.0 * float64(count)))

	return string(s)[:base] + strings.Repeat("*", len(s)-base)
}

func Secret(name string, value string) zap.Field {
	return zap.String(name, SecretString(value).String())
}
