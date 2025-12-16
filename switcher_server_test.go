package logging

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

func TestSwitcherServerHandler_BasicRequest(t *testing.T) {
	logger := zaptest.NewLogger(t)
	registry := newRegistry("test", logger)
	
	handler := &switcherServerHandler{
		registry:       registry,
		patternTracker: nil, // Test without pattern tracker first
	}

	req := logChangeReq{
		Inputs: "test.package",
		Level:  "DEBUG",
	}
	
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("PUT", "/", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, httpReq)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	if w.Body.String() != "ok" {
		t.Errorf("Expected body 'ok', got %q", w.Body.String())
	}
}

func TestSwitcherServerHandler_WithPatternTracker(t *testing.T) {
	logger := zaptest.NewLogger(t)
	registry := newRegistry("test", logger)
	tracker := newPatternTracker(registry, 5*time.Minute, logger)
	
	handler := &switcherServerHandler{
		registry:       registry,
		patternTracker: tracker,
	}

	// Test DEBUG level request (should be tracked)
	req := logChangeReq{
		Inputs:    "test.package",
		Level:     "DEBUG",
		Permanent: false,
	}
	
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("PUT", "/", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, httpReq)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	// Check that pattern was tracked
	if tracker.getActivePatternCount() != 1 {
		t.Errorf("Expected 1 tracked pattern, got %d", tracker.getActivePatternCount())
	}
	
	patterns := tracker.getActivePatterns()
	pattern, exists := patterns["test.package"]
	if !exists {
		t.Error("Expected pattern 'test.package' to be tracked")
	}
	
	if pattern.level != zapcore.DebugLevel {
		t.Errorf("Expected DEBUG level, got %v", pattern.level)
	}
	
	if pattern.permanent {
		t.Error("Expected permanent to be false")
	}
}

func TestSwitcherServerHandler_PermanentFlag(t *testing.T) {
	logger := zaptest.NewLogger(t)
	registry := newRegistry("test", logger)
	tracker := newPatternTracker(registry, 5*time.Minute, logger)
	
	handler := &switcherServerHandler{
		registry:       registry,
		patternTracker: tracker,
	}

	// Test DEBUG level request with permanent flag
	req := logChangeReq{
		Inputs:    "test.package",
		Level:     "DEBUG",
		Permanent: true,
	}
	
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("PUT", "/", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, httpReq)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	// Check that pattern was tracked as permanent
	patterns := tracker.getActivePatterns()
	pattern, exists := patterns["test.package"]
	if !exists {
		t.Error("Expected pattern 'test.package' to be tracked")
	}
	
	if !pattern.permanent {
		t.Error("Expected permanent to be true")
	}
}

func TestSwitcherServerHandler_TraceLevel(t *testing.T) {
	logger := zaptest.NewLogger(t)
	registry := newRegistry("test", logger)
	tracker := newPatternTracker(registry, 5*time.Minute, logger)
	
	handler := &switcherServerHandler{
		registry:       registry,
		patternTracker: tracker,
	}

	// Test TRACE level request
	req := logChangeReq{
		Inputs: "test.package",
		Level:  "TRACE",
	}
	
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("PUT", "/", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, httpReq)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	// Check that pattern was tracked with trace level
	patterns := tracker.getActivePatterns()
	pattern, exists := patterns["test.package"]
	if !exists {
		t.Error("Expected pattern 'test.package' to be tracked")
	}
	
	if pattern.level != zapcore.DebugLevel-1 {
		t.Errorf("Expected TRACE level (DEBUG-1), got %v", pattern.level)
	}
}

func TestSwitcherServerHandler_InfoLevelRemovesPattern(t *testing.T) {
	logger := zaptest.NewLogger(t)
	registry := newRegistry("test", logger)
	tracker := newPatternTracker(registry, 5*time.Minute, logger)
	
	handler := &switcherServerHandler{
		registry:       registry,
		patternTracker: tracker,
	}

	// First, add a DEBUG pattern
	req := logChangeReq{
		Inputs: "test.package",
		Level:  "DEBUG",
	}
	
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("PUT", "/", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, httpReq)
	
	if tracker.getActivePatternCount() != 1 {
		t.Error("Expected 1 pattern after DEBUG request")
	}

	// Now set to INFO level (should remove pattern)
	req.Level = "INFO"
	body, _ = json.Marshal(req)
	httpReq = httptest.NewRequest("PUT", "/", bytes.NewReader(body))
	w = httptest.NewRecorder()

	handler.ServeHTTP(w, httpReq)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	// Pattern should be removed
	if tracker.getActivePatternCount() != 0 {
		t.Error("Expected 0 patterns after INFO request")
	}
}

func TestSwitcherServerHandler_InvalidLevel(t *testing.T) {
	logger := zaptest.NewLogger(t)
	registry := newRegistry("test", logger)
	
	handler := &switcherServerHandler{
		registry:       registry,
		patternTracker: nil,
	}

	req := logChangeReq{
		Inputs: "test.package",
		Level:  "INVALID",
	}
	
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("PUT", "/", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, httpReq)

	if w.Code != 400 {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestSwitcherServerHandler_EmptyInputs(t *testing.T) {
	logger := zaptest.NewLogger(t)
	registry := newRegistry("test", logger)
	
	handler := &switcherServerHandler{
		registry:       registry,
		patternTracker: nil,
	}

	req := logChangeReq{
		Inputs: "",
		Level:  "DEBUG",
	}
	
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("PUT", "/", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, httpReq)

	if w.Code != 400 {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestSwitcherServerHandler_InvalidJSON(t *testing.T) {
	logger := zaptest.NewLogger(t)
	registry := newRegistry("test", logger)
	
	handler := &switcherServerHandler{
		registry:       registry,
		patternTracker: nil,
	}

	httpReq := httptest.NewRequest("PUT", "/", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, httpReq)

	if w.Code != 400 {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestSwitcherServerHandler_PatternUpdate(t *testing.T) {
	logger := zaptest.NewLogger(t)
	registry := newRegistry("test", logger)
	tracker := newPatternTracker(registry, 5*time.Minute, logger)
	
	handler := &switcherServerHandler{
		registry:       registry,
		patternTracker: tracker,
	}

	// First request: DEBUG, non-permanent
	req := logChangeReq{
		Inputs:    "test.package",
		Level:     "DEBUG",
		Permanent: false,
	}
	
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("PUT", "/", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, httpReq)
	
	patterns := tracker.getActivePatterns()
	initialTime := patterns["test.package"].timestamp

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	// Second request: TRACE, permanent (should update existing pattern)
	req.Level = "TRACE"
	req.Permanent = true
	
	body, _ = json.Marshal(req)
	httpReq = httptest.NewRequest("PUT", "/", bytes.NewReader(body))
	w = httptest.NewRecorder()

	handler.ServeHTTP(w, httpReq)

	// Should still have 1 pattern, but updated
	if tracker.getActivePatternCount() != 1 {
		t.Errorf("Expected 1 pattern after update, got %d", tracker.getActivePatternCount())
	}
	
	patterns = tracker.getActivePatterns()
	pattern := patterns["test.package"]
	
	if pattern.level != zapcore.DebugLevel-1 {
		t.Errorf("Expected TRACE level, got %v", pattern.level)
	}
	
	if !pattern.permanent {
		t.Error("Expected permanent to be true after update")
	}
	
	if !pattern.timestamp.After(initialTime) {
		t.Error("Expected timestamp to be updated")
	}
}
