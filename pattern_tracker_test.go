package logging

import (
	"context"
	"testing"
	"time"

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
	
	if len(patterns) != 1 {
		t.Errorf("Expected 1 pattern, got %d", len(patterns))
	}
	
	pattern, exists := patterns["test.debug"]
	if !exists {
		t.Error("Expected pattern 'test.debug' to exist")
	}
	
	if pattern.level != zapcore.DebugLevel {
		t.Errorf("Expected level DEBUG, got %v", pattern.level)
	}
	
	if pattern.permanent {
		t.Error("Expected permanent to be false")
	}

	// Test adding trace pattern (below debug level)
	tracker.addOrUpdatePattern("test.trace", zapcore.DebugLevel-1, true)
	patterns = tracker.getActivePatterns()
	
	if len(patterns) != 2 {
		t.Errorf("Expected 2 patterns, got %d", len(patterns))
	}
	
	tracePattern, exists := patterns["test.trace"]
	if !exists {
		t.Error("Expected pattern 'test.trace' to exist")
	}
	
	if !tracePattern.permanent {
		t.Error("Expected permanent to be true")
	}

	// Test that INFO level patterns are not tracked
	tracker.addOrUpdatePattern("test.info", zapcore.InfoLevel, false)
	patterns = tracker.getActivePatterns()
	
	if len(patterns) != 2 {
		t.Errorf("Expected 2 patterns (INFO should not be tracked), got %d", len(patterns))
	}
}

func TestPatternTracker_RemovePattern(t *testing.T) {
	logger := zaptest.NewLogger(t)
	registry := newRegistry("test", logger)
	tracker := newPatternTracker(registry, 5*time.Minute, logger)

	// Add a pattern
	tracker.addOrUpdatePattern("test.debug", zapcore.DebugLevel, false)
	if tracker.getActivePatternCount() != 1 {
		t.Error("Expected 1 pattern after adding")
	}

	// Remove the pattern
	tracker.removePattern("test.debug")
	if tracker.getActivePatternCount() != 0 {
		t.Error("Expected 0 patterns after removing")
	}

	// Removing non-existent pattern should not cause issues
	tracker.removePattern("non.existent")
	if tracker.getActivePatternCount() != 0 {
		t.Error("Expected 0 patterns after removing non-existent pattern")
	}
}

func TestPatternTracker_UpdateExistingPattern(t *testing.T) {
	logger := zaptest.NewLogger(t)
	registry := newRegistry("test", logger)
	tracker := newPatternTracker(registry, 5*time.Minute, logger)

	// Add initial pattern
	tracker.addOrUpdatePattern("test.debug", zapcore.DebugLevel, false)
	patterns := tracker.getActivePatterns()
	initialTime := patterns["test.debug"].timestamp

	// Wait a bit to ensure timestamp difference
	time.Sleep(10 * time.Millisecond)

	// Update the same pattern
	tracker.addOrUpdatePattern("test.debug", zapcore.DebugLevel-1, true)
	patterns = tracker.getActivePatterns()
	
	if len(patterns) != 1 {
		t.Errorf("Expected 1 pattern after update, got %d", len(patterns))
	}
	
	pattern := patterns["test.debug"]
	if pattern.level != zapcore.DebugLevel-1 {
		t.Errorf("Expected level to be updated to TRACE, got %v", pattern.level)
	}
	
	if !pattern.permanent {
		t.Error("Expected permanent to be updated to true")
	}
	
	if !pattern.timestamp.After(initialTime) {
		t.Error("Expected timestamp to be updated")
	}
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
	
	if tracker.getActivePatternCount() != 1 {
		t.Error("Expected 1 pattern before auto-reset")
	}
	
	// Wait for auto-reset to trigger
	time.Sleep(200 * time.Millisecond)
	
	// Manually trigger the check (since we're not running the background loop)
	tracker.checkAndResetExpiredPatterns()
	
	if tracker.getActivePatternCount() != 0 {
		t.Error("Expected 0 patterns after auto-reset")
	}
}

func TestPatternTracker_PermanentPatternNotReset(t *testing.T) {
	logger := zaptest.NewLogger(t)
	registry := newRegistry("test", logger)
	
	// Use very short timeout for testing
	tracker := newPatternTracker(registry, 100*time.Millisecond, logger)
	
	// Add a permanent pattern
	tracker.addOrUpdatePattern("test", zapcore.DebugLevel, true)
	
	if tracker.getActivePatternCount() != 1 {
		t.Error("Expected 1 pattern before timeout")
	}
	
	// Wait for timeout period
	time.Sleep(200 * time.Millisecond)
	
	// Manually trigger the check
	tracker.checkAndResetExpiredPatterns()
	
	// Permanent pattern should still be there
	if tracker.getActivePatternCount() != 1 {
		t.Error("Expected 1 pattern after timeout (permanent should not be reset)")
	}
}

func TestPatternTracker_StartStop(t *testing.T) {
	logger := zaptest.NewLogger(t)
	registry := newRegistry("test", logger)
	tracker := newPatternTracker(registry, 5*time.Minute, logger)

	// Start the tracker
	tracker.start()
	
	// Verify it's running by checking context
	select {
	case <-tracker.ctx.Done():
		t.Error("Context should not be done immediately after start")
	default:
		// Good, context is not done
	}
	
	// Stop the tracker
	done := make(chan struct{})
	go func() {
		tracker.stop()
		close(done)
	}()
	
	// Should complete within reasonable time
	select {
	case <-done:
		// Good, stop completed
	case <-time.After(5 * time.Second):
		t.Error("Stop did not complete within 5 seconds")
	}
	
	// Context should be done after stop
	select {
	case <-tracker.ctx.Done():
		// Good, context is done
	default:
		t.Error("Context should be done after stop")
	}
}

func TestPatternTracker_ConcurrentAccess(t *testing.T) {
	logger := zaptest.NewLogger(t)
	registry := newRegistry("test", logger)
	tracker := newPatternTracker(registry, 5*time.Minute, logger)

	// Test concurrent adds and removes
	done := make(chan struct{})
	
	// Goroutine 1: Add patterns
	go func() {
		for i := 0; i < 100; i++ {
			tracker.addOrUpdatePattern("pattern1", zapcore.DebugLevel, false)
			time.Sleep(time.Microsecond)
		}
		done <- struct{}{}
	}()
	
	// Goroutine 2: Remove patterns
	go func() {
		for i := 0; i < 100; i++ {
			tracker.removePattern("pattern1")
			time.Sleep(time.Microsecond)
		}
		done <- struct{}{}
	}()
	
	// Goroutine 3: Read patterns
	go func() {
		for i := 0; i < 100; i++ {
			_ = tracker.getActivePatterns()
			time.Sleep(time.Microsecond)
		}
		done <- struct{}{}
	}()
	
	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		select {
		case <-done:
			// Good
		case <-time.After(5 * time.Second):
			t.Error("Concurrent access test did not complete within 5 seconds")
			return
		}
	}
	
	// Should not panic or deadlock
}
