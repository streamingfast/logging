package zapx

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHumanDuration_String(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		// Milliseconds (< 0.1s)
		{"1ms", 1 * time.Millisecond, "1.00ms"},
		{"50ms", 50 * time.Millisecond, "50.00ms"},
		{"99.5ms", 99500 * time.Microsecond, "99.50ms"},

		// Seconds (>= 0.1s and < 60s)
		{"0.1s", 100 * time.Millisecond, "0.1s"},
		{"0.5s", 500 * time.Millisecond, "0.5s"},
		{"1s", 1 * time.Second, "1s"},
		{"5.5s", 5*time.Second + 500*time.Millisecond, "5.5s"},
		{"10.25s", 10*time.Second + 250*time.Millisecond, "10.25s"},
		{"59.99s", 59*time.Second + 990*time.Millisecond, "59.99s"},

		// Minutes + seconds (>= 60s and < 120m)
		{"1m 0s", 1 * time.Minute, "1m 0s"},
		{"1m 30.5s", 1*time.Minute + 30*time.Second + 500*time.Millisecond, "1m 30.5s"},
		{"30m 45.55s", 30*time.Minute + 45*time.Second + 550*time.Millisecond, "30m 45.55s"},
		{"119m 59.99s", 119*time.Minute + 59*time.Second + 990*time.Millisecond, "119m 59.99s"},

		// Hours + minutes + seconds (>= 120m)
		{"2h 0m 0s", 2 * time.Hour, "2h 0m 0s"},
		{"2h 5m 10.5s", 2*time.Hour + 5*time.Minute + 10*time.Second + 500*time.Millisecond, "2h 5m 10.5s"},
		{"20h 40m 54.34s", 20*time.Hour + 40*time.Minute + 54*time.Second + 340*time.Millisecond, "20h 40m 54.34s"},
		{"100h 0m 0s", 100 * time.Hour, "100h 0m 0s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hd := humanDuration(tt.duration)
			result := hd.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}
