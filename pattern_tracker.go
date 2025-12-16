package logging

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// activePattern represents a debug/trace pattern that is being tracked for auto-reset
type activePattern struct {
	originalInput string
	level         zapcore.Level
	timestamp     time.Time
	permanent     bool
}

// patternTracker manages active debug/trace patterns and handles auto-reset functionality
type patternTracker struct {
	mu               sync.RWMutex
	activePatterns   map[string]*activePattern
	autoResetTimeout time.Duration
	registry         *registry
	logger           *zap.Logger
	ctx              context.Context
	cancel           context.CancelFunc
	done             chan struct{}
}

// newPatternTracker creates a new pattern tracker with the specified auto-reset timeout
func newPatternTracker(registry *registry, autoResetTimeout time.Duration, logger *zap.Logger) *patternTracker {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &patternTracker{
		activePatterns:   make(map[string]*activePattern),
		autoResetTimeout: autoResetTimeout,
		registry:         registry,
		logger:           logger,
		ctx:              ctx,
		cancel:           cancel,
		done:             make(chan struct{}),
	}
}

// start begins the background goroutine that periodically checks for expired patterns
func (pt *patternTracker) start() {
	go pt.resetLoop()
}

// stop gracefully shuts down the pattern tracker
func (pt *patternTracker) stop() {
	pt.cancel()
	<-pt.done
}

// addOrUpdatePattern adds a new pattern or updates an existing one
func (pt *patternTracker) addOrUpdatePattern(input string, level zapcore.Level, permanent bool) {
	// Only track debug and trace levels for auto-reset
	if level != zapcore.DebugLevel && level < zapcore.DebugLevel {
		pt.mu.Lock()
		defer pt.mu.Unlock()
		
		pt.activePatterns[input] = &activePattern{
			originalInput: input,
			level:         level,
			timestamp:     time.Now(),
			permanent:     permanent,
		}
		
		pt.logger.Debug("added/updated pattern for auto-reset tracking",
			zap.String("pattern", input),
			zap.Stringer("level", level),
			zap.Bool("permanent", permanent),
		)
	}
}

// removePattern removes a pattern from tracking (when set to higher level)
// This removes both permanent and non-permanent patterns
func (pt *patternTracker) removePattern(input string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	
	if pattern, exists := pt.activePatterns[input]; exists {
		delete(pt.activePatterns, input)
		pt.logger.Debug("removed pattern from auto-reset tracking", 
			zap.String("pattern", input),
			zap.Bool("was_permanent", pattern.permanent))
	}
}

// resetLoop is the background goroutine that periodically checks for expired patterns
func (pt *patternTracker) resetLoop() {
	defer close(pt.done)
	
	// Check every minute for expired patterns
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-pt.ctx.Done():
			pt.logger.Debug("pattern tracker reset loop shutting down")
			return
		case <-ticker.C:
			pt.checkAndResetExpiredPatterns()
		}
	}
}

// checkAndResetExpiredPatterns scans for expired patterns and resets them to INFO level
func (pt *patternTracker) checkAndResetExpiredPatterns() {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	
	now := time.Now()
	var expiredPatterns []string
	
	for input, pattern := range pt.activePatterns {
		// Skip permanent patterns
		if pattern.permanent {
			continue
		}
		
		// Check if pattern has expired
		if now.Sub(pattern.timestamp) >= pt.autoResetTimeout {
			expiredPatterns = append(expiredPatterns, input)
		}
	}
	
	// Reset expired patterns to INFO level
	for _, input := range expiredPatterns {
		pattern := pt.activePatterns[input]
		pt.logger.Info("auto-resetting expired debug/trace pattern to INFO level",
			zap.String("pattern", input),
			zap.Stringer("original_level", pattern.level),
			zap.Duration("elapsed", now.Sub(pattern.timestamp)),
		)
		
		// Create a spec to reset this pattern to INFO level
		spec := newLogLevelSpec(envGetFromMap(map[string]string{
			"INFO": input,
		}))
		
		// Apply the reset to all matching loggers
		pt.registry.forAllEntriesMatchingSpec(spec, func(entry *registryEntry, level zapcore.Level, trace bool) {
			pt.registry.setLevelForEntry(entry, zapcore.InfoLevel, false)
		})
		
		// Remove from active patterns
		delete(pt.activePatterns, input)
	}
	
	if len(expiredPatterns) > 0 {
		pt.logger.Info("completed auto-reset of expired patterns", zap.Int("reset_count", len(expiredPatterns)))
	}
}

// getActivePatternCount returns the number of currently tracked patterns (for testing/monitoring)
func (pt *patternTracker) getActivePatternCount() int {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	return len(pt.activePatterns)
}

// getActivePatterns returns a copy of currently tracked patterns (for testing/monitoring)
func (pt *patternTracker) getActivePatterns() map[string]activePattern {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	
	result := make(map[string]activePattern)
	for k, v := range pt.activePatterns {
		result[k] = *v
	}
	return result
}
