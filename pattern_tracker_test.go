package logging

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

func TestPatternTracker_AddOrUpdatePattern(t *testing.T) {
	logger := zaptest.NewLogger(t)
	registry := newRegistry("test", logger)
	tracker := newPatternTracker(registry, 5*time.Minute, logger)

	// Test adding debug pattern
	tracker.addOrUpdatePattern("test.debug", zapcore.DebugLevel, false)
	patterns := tracker.getActivePatterns()
	require.Len(t, patterns, 1)

	pattern, exists := patterns["test.debug"]
	require.True(t, exists, "expected pattern 'test.debug' to exist")
	assert.Equal(t, zapcore.DebugLevel, pattern.level)
	assert.False(t, pattern.permanent)

	// Test adding trace pattern (below debug level)
	tracker.addOrUpdatePattern("test.trace", zapcore.DebugLevel-1, true)
	patterns = tracker.getActivePatterns()
	require.Len(t, patterns, 2)

	tracePattern, exists := patterns["test.trace"]
	require.True(t, exists, "expected pattern 'test.trace' to exist")
	assert.True(t, tracePattern.permanent)

	// Test that INFO level patterns are not tracked
	tracker.addOrUpdatePattern("test.info", zapcore.InfoLevel, false)
	assert.Len(t, tracker.getActivePatterns(), 2, "INFO should not be tracked")
}

func TestPatternTracker_RemovePattern(t *testing.T) {
	logger := zaptest.NewLogger(t)
	registry := newRegistry("test", logger)
	tracker := newPatternTracker(registry, 5*time.Minute, logger)

	// Add a pattern
	tracker.addOrUpdatePattern("test.debug", zapcore.DebugLevel, false)
	assert.Equal(t, 1, tracker.getActivePatternCount(), "after adding")

	// Remove the pattern
	tracker.removePattern("test.debug")
	assert.Equal(t, 0, tracker.getActivePatternCount(), "after removing")

	// Removing non-existent pattern should not cause issues
	tracker.removePattern("non.existent")
	assert.Equal(t, 0, tracker.getActivePatternCount(), "after removing non-existent pattern")
}

func TestPatternTracker_UpdateExistingPattern(t *testing.T) {
	logger := zaptest.NewLogger(t)
	registry := newRegistry("test", logger)
	tracker := newPatternTracker(registry, 5*time.Minute, logger)

	// Add initial pattern
	tracker.addOrUpdatePattern("test.debug", zapcore.DebugLevel, false)
	initialTime := tracker.getActivePatterns()["test.debug"].timestamp

	// Wait a bit to ensure timestamp difference
	time.Sleep(10 * time.Millisecond)

	// Update the same pattern
	tracker.addOrUpdatePattern("test.debug", zapcore.DebugLevel-1, true)
	patterns := tracker.getActivePatterns()
	require.Len(t, patterns, 1)

	pattern := patterns["test.debug"]
	assert.Equal(t, zapcore.DebugLevel-1, pattern.level, "level should be updated to TRACE")
	assert.True(t, pattern.permanent, "permanent should be updated to true")
	assert.True(t, pattern.timestamp.After(initialTime), "timestamp should be updated")
}

func TestPatternTracker_AutoReset(t *testing.T) {
	logger := zaptest.NewLogger(t)
	registry := newRegistry("test", logger)

	// Create a test logger entry
	testLogger := zap.NewNop()
	register(registry, "test.package", testLogger, loggerShortName("test"))

	// Use very short timeout for testing
	tracker := newPatternTracker(registry, 100*time.Millisecond, logger)

	// Add a non-permanent pattern
	tracker.addOrUpdatePattern("test", zapcore.DebugLevel, false)
	assert.Equal(t, 1, tracker.getActivePatternCount(), "before auto-reset")

	// Wait for auto-reset to trigger
	time.Sleep(200 * time.Millisecond)

	// Manually trigger the check (since we're not running the background loop)
	tracker.checkAndResetExpiredPatterns()
	assert.Equal(t, 0, tracker.getActivePatternCount(), "after auto-reset")
}

func TestPatternTracker_PermanentPatternNotReset(t *testing.T) {
	logger := zaptest.NewLogger(t)
	registry := newRegistry("test", logger)

	// Use very short timeout for testing
	tracker := newPatternTracker(registry, 100*time.Millisecond, logger)

	// Add a permanent pattern
	tracker.addOrUpdatePattern("test", zapcore.DebugLevel, true)
	assert.Equal(t, 1, tracker.getActivePatternCount(), "before timeout")

	// Wait for timeout period
	time.Sleep(200 * time.Millisecond)

	// Manually trigger the check
	tracker.checkAndResetExpiredPatterns()

	// Permanent pattern should still be there
	assert.Equal(t, 1, tracker.getActivePatternCount(), "permanent pattern should not be reset")
}

func TestPatternTracker_StartStop(t *testing.T) {
	logger := zaptest.NewLogger(t)
	registry := newRegistry("test", logger)
	tracker := newPatternTracker(registry, 5*time.Minute, logger)

	// Start the tracker
	tracker.start()

	// Verify it's running by checking context
	assert.NoError(t, tracker.ctx.Err(), "context should not be done immediately after start")

	// Stop the tracker
	done := make(chan struct{})
	go func() {
		tracker.stop()
		close(done)
	}()

	// Should complete within reasonable time
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		require.Fail(t, "stop did not complete within 5 seconds")
	}

	// Context should be done after stop
	assert.Error(t, tracker.ctx.Err(), "context should be done after stop")
}

func TestPatternTracker_ConcurrentAccess(t *testing.T) {
	logger := zaptest.NewLogger(t)
	registry := newRegistry("test", logger)
	tracker := newPatternTracker(registry, 5*time.Minute, logger)

	var wg sync.WaitGroup

	// Add patterns
	wg.Go(func() {
		for range 100 {
			tracker.addOrUpdatePattern("pattern1", zapcore.DebugLevel, false)
			time.Sleep(time.Microsecond)
		}
	})

	// Remove patterns
	wg.Go(func() {
		for range 100 {
			tracker.removePattern("pattern1")
			time.Sleep(time.Microsecond)
		}
	})

	// Read patterns
	wg.Go(func() {
		for range 100 {
			_ = tracker.getActivePatterns()
			time.Sleep(time.Microsecond)
		}
	})

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// Should not panic or deadlock
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		require.Fail(t, "concurrent access test did not complete within 5 seconds")
	}
}
