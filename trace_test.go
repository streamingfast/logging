package logging

import (
	"os"
	"strings"
	"testing"

	"github.com/test-go/testify/assert"
)

func TestIsTraceEnabled(t *testing.T) {
	tests := []struct {
		name            string
		inName          string
		inID            string
		env             string
		expectedEnabled bool
	}{
		{"match name", "log", "github/log", "TRACE=log", true},
		{"match id", "log", "github/log", "TRACE=github/log", true},
		{"match id regex", "log", "github/log", "TRACE=github/.*", true},
		{"match all", "log", "github/log", "TRACE=.*", true},

		{"match then denied by name", "log", "github/log", "TRACE=.*,-log", false},
		{"match then denied by id", "log", "github/log", "TRACE=.*,-github/log", false},
		{"match then denied by id regex", "log", "github/log", "TRACE=.*,-github/.*", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			os.Setenv("TRACE", strings.TrimPrefix(test.env, "TRACE="))
			isEnabled := IsTraceEnabled(test.inName, test.inID)
			assert.Equal(t, test.expectedEnabled, isEnabled, "Expected trace to be enabled=%t for %s @ %s with env %s but it was enabled=%t", test.expectedEnabled, test.inName, test.inID, test.env, isEnabled)
		})
	}
}
