package zapx

import (
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

// HumanDuration prints the duration as compact humanize format where the unit
// is present and where decimals are used for precision. The actual precision is always
// truncated to 2 digits so this logging loses precisions in most cases.
//
// - If the duration is smaller than 0.1s, prints in milliseconds with 2 digits precision.
// - If the duration is smaller than 60s but bigger than 0.1s, prints in seconds with 2 digits
// precision with `s` as suffix.
// - If the duration is smaller than 120m but bigger than 60s, prints in minutes + seconds with
// 2 digits precision (e.g. 30m 45.55s).
// - Otherwise, prints as hour + minutes + seconds with 2 digits precision (e.g. 20h 40m 54.34s).
//
// Trailing zeros as not printed from the decimal portion is present.
func HumanDuration(name string, duration time.Duration) zap.Field {
	return zap.Stringer(name, humanDuration(duration))
}

type humanDuration time.Duration

func (d humanDuration) String() string {
	duration := time.Duration(d)

	// If the duration is smaller than 0.1s, prints in milliseconds with 2 digits precision
	if duration < 100*time.Millisecond {
		ms := float64(duration) / float64(time.Millisecond)
		return fmt.Sprintf("%.2fms", ms)
	}

	// If the duration is smaller than 60s but bigger than 0.1s, prints in seconds with 2 digits precision
	if duration < 60*time.Second {
		s := float64(duration) / float64(time.Second)
		formatted := fmt.Sprintf("%.2f", s)
		// Trim trailing zeros from decimal portion
		formatted = strings.TrimRight(formatted, "0")
		formatted = strings.TrimRight(formatted, ".")
		return formatted + "s"
	}

	// If the duration is smaller than 120m but bigger than 60s, prints in minutes + seconds with 2 digits precision
	if duration < 120*time.Minute {
		minutes := int(duration / time.Minute)
		remaining := duration - time.Duration(minutes)*time.Minute
		s := float64(remaining) / float64(time.Second)
		secStr := strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", s), "0"), ".")
		return fmt.Sprintf("%dm %ss", minutes, secStr)
	}

	// Otherwise, prints as hour + minutes + seconds with 2 digits precision
	hours := int(duration / time.Hour)
	remaining := duration - time.Duration(hours)*time.Hour
	minutes := int(remaining / time.Minute)
	remaining = remaining - time.Duration(minutes)*time.Minute
	s := float64(remaining) / float64(time.Second)
	secStr := strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", s), "0"), ".")
	return fmt.Sprintf("%dh %dm %ss", hours, minutes, secStr)
}
