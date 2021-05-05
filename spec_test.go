package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestSortedEntriesFromEnv(t *testing.T) {
	tests := []struct {
		name     string
		in       func(string) string
		expected []*levelSpec
	}{
		{
			"standard",
			fakeEnv(map[string]string{
				"DEBUG": "true",
			}),
			[]*levelSpec{
				{key: "true", level: zapcore.DebugLevel, trace: false, ordering: 1},
			},
		},
		{
			"debug=true and trace=true",
			fakeEnv(map[string]string{
				"DEBUG": "true",
				"TRACE": "true",
			}),
			[]*levelSpec{
				{key: "true", level: zapcore.DebugLevel, trace: false, ordering: 1},
				{key: "true", level: zapcore.DebugLevel, trace: true, ordering: 2},
			},
		},
		{
			"trace=*",
			fakeEnv(map[string]string{
				"TRACE": "*",
			}),
			[]*levelSpec{
				{key: "*", level: zapcore.DebugLevel, trace: true, ordering: 1},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, newLogLevelSpec(test.in).sortedSpecs())
		})
	}
}
