package logging

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"go.uber.org/zap/zapcore"
)

type logChangeReq struct {
	Inputs    string `json:"inputs"`
	Level     string `json:"level"`
	Permanent bool   `json:"permanent,omitempty"`
}

type switcherServerHandler struct {
	registry       *registry
	patternTracker *patternTracker
}

func (h *switcherServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	in := logChangeReq{}
	if err := decoder.Decode(&in); err != nil {
		http.Error(w, fmt.Sprintf("cannot unmarshal request: %s", err), 400)
		return
	}

	if in.Inputs == "" {
		http.Error(w, "inputs not defined, should be comma-separated list of words or a regular expressions", 400)
		return
	}

	level := strings.ToUpper(in.Level)
	if level != "WARN" && level != "WARNING" && level != "INFO" && level != "DEBUG" && level != "TRACE" {
		http.Error(w, fmt.Sprintf("invalid level value %q", in.Level), 400)
		return
	}

	spec := newLogLevelSpec(envGetFromMap(map[string]string{
		strings.ToUpper(in.Level): in.Inputs,
	}))

	// Convert string level to zapcore.Level for pattern tracking
	var zapLevel zapcore.Level
	switch level {
	case "TRACE":
		zapLevel = zapcore.DebugLevel - 1 // Trace is below debug
	case "DEBUG":
		zapLevel = zapcore.DebugLevel
	case "INFO":
		zapLevel = zapcore.InfoLevel
	case "WARN", "WARNING":
		zapLevel = zapcore.WarnLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	// Handle pattern tracking for auto-reset functionality
	if h.patternTracker != nil {
		if zapLevel == zapcore.DebugLevel || zapLevel < zapcore.DebugLevel {
			// Add debug/trace patterns to tracking
			h.patternTracker.addOrUpdatePattern(in.Inputs, zapLevel, in.Permanent)
		} else {
			// Remove patterns when set to higher levels (INFO, WARN, ERROR)
			h.patternTracker.removePattern(in.Inputs)
		}
	}

	globalRegistry.forAllEntriesMatchingSpec(spec, func(entry *registryEntry, level zapcore.Level, trace bool) {
		globalRegistry.setLevelForEntry(entry, level, trace)
	})

	w.Write([]byte("ok"))
}
